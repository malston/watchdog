import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import _ from 'lodash';

const SimpleMonitor = () => {
  const [connectionData, setConnectionData] = useState([]);
  const [setCurrentStatus] = useState({ status: 'UNKNOWN', since: '', duration: '' });
  const [lastUpdated, setLastUpdated] = useState(null);
  const [stats, setStats] = useState({
    uptime: 0,
    downtime: 0,
    changes: 0,
    avgLatency: 0,
    maxLatency: 0,
    lastLatency: 0
  });

  // Mock function to simulate reading log file
  const fetchData = async () => {
    try {
          const response = await fetch('http://localhost:8080/api/connection-data');
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          
          const data = await response.json();
          
          // Process the data
          const processedData = data.map(entry => ({
            ...entry,
            timestamp: new Date(entry.timestamp),
            latency: entry.status === 'DOWN' ? 0 : parseInt(entry.latency),
            isDown: entry.status === 'DOWN'
          }));
          
          setConnectionData(processedData);
          
          // Update the current status (same as before)
          if (processedData.length > 0) {
            const latest = processedData[processedData.length - 1];
            setCurrentStatus({
              status: latest.status,
              since: latest.timestamp.toLocaleTimeString(),
              duration: latest.status === 'UP' ? latest.uptime : latest.downtime
            });
            
            // Calculate stats
            const upRecords = processedData.filter(d => d.status === 'UP');
            setStats({
              uptime: processedData.length > 0 ? latest.uptime : '0s',
              downtime: processedData.length > 0 ? latest.downtime : '0s',
              changes: processedData.length > 0 ? parseInt(latest.total_changes) : 0,
              avgLatency: upRecords.length > 0 ? Math.round(_.meanBy(upRecords, 'latency')) : 0,
              maxLatency: upRecords.length > 0 ? _.maxBy(upRecords, 'latency').latency : 0,
              lastLatency: latest.status === 'UP' ? parseInt(latest.latency) : 0
            });
          }
          
        } catch (error) {
          console.error("Error fetching connection data:", error);
          // Keep existing data if there was an error
        }

    // setConnectionData(processedData);
    setLastUpdated(new Date().toLocaleTimeString());
  };

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
        <h2 style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '10px' }}>Current Status</h2>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <div style={{ 
            width: '12px', 
            height: '12px', 
            borderRadius: '50%', 
            backgroundColor: statusColor[currentStatus],
            marginRight: '10px'
          }}></div>
          <span style={{ fontSize: '20px', fontWeight: 'bold' }}>{currentStatus}</span>
        </div>
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
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default SimpleMonitor;