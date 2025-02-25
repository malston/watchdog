import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import _ from 'lodash';

const SimpleMonitor = () => {
  const [connectionData, setConnectionData] = useState([]);
  const [lastUpdated, setLastUpdated] = useState(null);

  // Fetch data from the Go backend
  const fetchData = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/connection-data');

      if (!response.ok) {
        throw new Error(`Failed to fetch data: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();

      if (!Array.isArray(data) || data.length === 0) {
        console.log("No data available yet or empty data returned");
        return;
      }

      // Process the data - convert to proper types
      const processedData = data.map(entry => ({
        timestamp: new Date(entry.timestamp),
        status: entry.status,
        latency: entry.status === 'DOWN' ? 0 : parseInt(entry.latency || '0'),
        uptime: entry.uptime || '0s',
        downtime: entry.downtime || '0s',
        total_changes: parseInt(entry.total_changes || '0'),
        message: entry.message || '',
        isDown: entry.status === 'DOWN'
      }));

      setConnectionData(processedData);
      setLastUpdated(new Date().toLocaleTimeString());

      console.log("Data fetched successfully:", processedData);
    } catch (error) {
      console.error("Error fetching connection data:", error);
      // Don't update state on error to keep existing data
    }
  };

  // Calculate time duration in a human-readable format
  const calculateDuration = (fromTime) => {
    if (!statusStartTime) return "just started";

    const now = new Date();
    const diff = now - statusStartTime;

    const seconds = Math.floor(diff / 1000);
    if (seconds < 60) return `${seconds} seconds`;

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes} minute${minutes > 1 ? 's' : ''} ${seconds % 60} second${seconds % 60 !== 1 ? 's' : ''}`;

    const hours = Math.floor(minutes / 60);
    return `${hours} hour${hours > 1 ? 's' : ''} ${minutes % 60} minute${minutes % 60 !== 1 ? 's' : ''}`;
  };

  // Track status changes
  useEffect(() => {
    if (connectionData.length > 0) {
      const lastEntry = connectionData[connectionData.length - 1];

      // If this is our first data or status has changed, update the status start time
      if (!statusStartTime ||
          (connectionData.length > 1 &&
           connectionData[connectionData.length - 2].status !== lastEntry.status)) {
        setStatusStartTime(new Date());
      }
    }
  }, [connectionData]);

  // Fetch data on initial load
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  const statusColor = {
    UP: "green",
    DOWN: "red",
    UNKNOWN: "gray"
  };

  const currentStatus = connectionData.length > 0
    ? connectionData[connectionData.length - 1].status
    : "UNKNOWN";

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h1 style={{ fontSize: '24px', fontWeight: 'bold', marginBottom: '10px' }}>
        Watchdog: Internet Connection Monitor
      </h1>
      <p style={{ marginBottom: '20px' }}>Last updated: {lastUpdated}</p>

      {/* Current Status */}
      <div style={{
        padding: '20px',
        marginBottom: '20px',
        border: `1px solid ${statusColor[currentStatus]}`,
        borderRadius: '8px',
        backgroundColor: `${statusColor[currentStatus]}20`
      }}>
        <h2 style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '10px', textAlign: 'center' }}>Current Status</h2>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div style={{
            width: '12px',
            height: '12px',
            borderRadius: '50%',
            backgroundColor: statusColor[currentStatus],
            marginRight: '10px'
          }}></div>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>{currentStatus}</span>
        </div>
        {connectionData.length > 0 && (
          <div style={{ marginTop: '10px', fontWeight: 'normal', fontSize: '14px', textAlign: 'center' }}>
            <p>
              {currentStatus === 'UP' ? 'Uptime: ' : 'Downtime: '}
              <span style={{ fontWeight: 'bold' }}>
                {currentStatus === 'UP'
                  ? connectionData[connectionData.length - 1].uptime
                  : connectionData[connectionData.length - 1].downtime}
              </span>
            </p>
          </div>
        )}
      </div>

      {/* Latency Chart */}
      <div style={{
        padding: '20px',
        marginBottom: '20px',
        border: '1px solid #ddd',
        borderRadius: '8px',
        backgroundColor: 'white'
      }}>
        <h2 style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '10px' }}>Connection Latency</h2>
        <div style={{ height: '300px' }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={connectionData}
              margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="timestamp"
                tickFormatter={(timestamp) => new Date(timestamp).toLocaleTimeString()}
              />
              <YAxis />
              <Tooltip
                labelFormatter={(timestamp) => new Date(timestamp).toLocaleTimeString()}
                formatter={(value) => [`${value} ms`, 'Latency']}
              />
              <Line
                type="monotone"
                dataKey="latency"
                stroke="#3B82F6"
                dot={{ r: 4 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Event Log */}
      <div style={{
        padding: '20px',
        border: '1px solid #ddd',
        borderRadius: '8px',
        backgroundColor: 'white'
      }}>
        <h2 style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '10px' }}>Connection Events</h2>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ backgroundColor: '#f3f4f6' }}>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '1px solid #ddd' }}>Time</th>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '1px solid #ddd' }}>Status</th>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '1px solid #ddd' }}>Latency</th>
              <th style={{ padding: '10px', textAlign: 'left', borderBottom: '1px solid #ddd' }}>Message</th>
            </tr>
          </thead>
          <tbody>
            {connectionData.slice().reverse().map((entry, index) => (
              <tr key={index}>
                <td style={{ padding: '10px', borderBottom: '1px solid #ddd' }}>
                  {entry.timestamp.toLocaleTimeString()}
                </td>
                <td style={{ padding: '10px', borderBottom: '1px solid #ddd' }}>
                  <span style={{
                    padding: '3px 8px',
                    borderRadius: '12px',
                    backgroundColor: entry.status === 'UP' ? '#dcfce7' : '#fee2e2',
                    color: entry.status === 'UP' ? '#166534' : '#991b1b',
                    fontSize: '12px',
                    fontWeight: 'bold'
                  }}>
                    {entry.status}
                  </span>
                </td>
                <td style={{ padding: '10px', borderBottom: '1px solid #ddd' }}>
                  {entry.status === 'UP' ? `${entry.latency} ms` : '-'}
                </td>
                <td style={{ padding: '10px', borderBottom: '1px solid #ddd', textAlign: 'left' }}>
                  {entry.message}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default SimpleMonitor;
