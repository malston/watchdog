import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ReferenceLine, ResponsiveContainer } from 'recharts';
import _ from 'lodash';

const ConnectionMonitor = () => {
  const [connectionData, setConnectionData] = useState([]);
  const [currentStatus, setCurrentStatus] = useState({ status: 'UNKNOWN', since: '', duration: '' });
  const [stats, setStats] = useState({
    uptime: 0,
    downtime: 0,
    changes: 0,
    avgLatency: 0,
    maxLatency: 0,
    lastLatency: 0
  });
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState(null);

  const fetchData = async () => {
    setLoading(true);
    
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
    
    setLastUpdated(new Date().toLocaleTimeString());
    setLoading(false);
  };

  // Fetch data on initial load and periodically
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  if (loading && connectionData.length === 0) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="text-xl mb-4">Loading connection data...</div>
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">Watchdog: Internet Connection Monitor</h1>
        <div className="text-sm text-gray-500">Last updated: {lastUpdated}</div>
      </div>
      
      {/* Current status card */}
      <div className={`p-4 mb-6 rounded-lg border ${currentStatus.status === 'UP' ? 'bg-green-50 border-green-200' : currentStatus.status === 'DOWN' ? 'bg-red-50 border-red-200' : 'bg-gray-50 border-gray-200'}`}>
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold mb-1">Current Status</h2>
            <div className="flex items-center space-x-2">
              <div className={`h-4 w-4 rounded-full ${currentStatus.status === 'UP' ? 'bg-green-500' : currentStatus.status === 'DOWN' ? 'bg-red-500' : 'bg-gray-500'}`}></div>
              <div className="text-xl font-bold">{currentStatus.status}</div>
            </div>
            <div className="mt-1 text-sm">
              Since {currentStatus.since} ({currentStatus.duration})
            </div>
          </div>
          <div className="text-right">
            <div className="text-3xl font-bold">{stats.lastLatency} ms</div>
            <div className="text-sm text-gray-500">Current latency</div>
          </div>
        </div>
      </div>
      
      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
          <div className="text-sm text-gray-500">Uptime</div>
          <div className="text-xl font-bold">{stats.uptime}</div>
        </div>
        <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
          <div className="text-sm text-gray-500">Downtime</div>
          <div className="text-xl font-bold">{stats.downtime}</div>
        </div>
        <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
          <div className="text-sm text-gray-500">Connection Changes</div>
          <div className="text-xl font-bold">{stats.changes}</div>
        </div>
        <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
          <div className="text-sm text-gray-500">Avg Latency</div>
          <div className="text-xl font-bold">{stats.avgLatency} ms</div>
        </div>
      </div>
      
      {/* Latency graph */}
      <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm mb-6">
        <h2 className="text-lg font-semibold mb-4">Connection Latency</h2>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={connectionData}
              margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="timestamp" 
                tickFormatter={(timestamp) => new Date(timestamp).toLocaleTimeString()} 
                label={{ value: 'Time', position: 'insideBottomRight', offset: 0 }} 
              />
              <YAxis 
                label={{ value: 'Latency (ms)', angle: -90, position: 'insideLeft' }} 
                domain={[0, 'dataMax + 5']}
              />
              <Tooltip 
                labelFormatter={(timestamp) => new Date(timestamp).toLocaleTimeString()} 
                formatter={(value, name) => [value + ' ms', 'Latency']}
              />
              <Line 
                type="monotone" 
                dataKey="latency" 
                stroke="#3B82F6" 
                strokeWidth={2}
                dot={{ r: 4 }}
                activeDot={{ r: 6 }}
                name="Latency" 
              />
              {connectionData.map((entry, index) => 
                entry.isDown && (
                  <ReferenceLine 
                    key={index} 
                    x={entry.timestamp} 
                    stroke="red" 
                    strokeDasharray="3 3" 
                    strokeWidth={2}
                  />
                )
              )}
            </LineChart>
          </ResponsiveContainer>
        </div>
        <div className="mt-2 text-sm text-gray-500 text-center">
          <span className="inline-block mx-2">
            <span className="inline-block w-3 h-3 bg-blue-500 rounded-full mr-1"></span> Latency
          </span>
          <span className="inline-block mx-2">
            <span className="inline-block w-3 h-3 bg-red-500 rounded-full mr-1"></span> Connection Drop
          </span>
        </div>
      </div>
      
      {/* Connection status timeline */}
      <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm">
        <h2 className="text-lg font-semibold mb-4">Connection Status Timeline</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Latency</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uptime</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {connectionData.slice().reverse().map((entry, index) => (
                <tr key={index}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {entry.timestamp.toLocaleTimeString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                      entry.status === 'UP' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {entry.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {entry.status === 'UP' ? `${entry.latency} ms` : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {entry.status === 'UP' ? `${entry.uptime}` : '-'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {entry.message}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default ConnectionMonitor;
