'use client'

import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface ConnectionEntry {
  timestamp: Date;
  status: 'UP' | 'DOWN' | 'UNKNOWN';
  latency: number;
  uptime: string;
  downtime: string;
  total_changes: number;
  message: string;
  isDown: boolean;
}

const SimpleMonitor: React.FC = () => {
  const [connectionData, setConnectionData] = useState<ConnectionEntry[]>([]);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  const [statusStartTime, setStatusStartTime] = useState<Date | null>(null);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [eventsPerPage] = useState<number>(10);
  const [totalEvents, setTotalEvents] = useState<number>(0);

  // Fetch data from the Go backend
  const fetchData = async (): Promise<void> => {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/connection-data`);

      if (!response.ok) {
        throw new Error(`Failed to fetch data: ${response.status} ${response.statusText}`);
      }

      const data = await response.json();

      if (!Array.isArray(data) || data.length === 0) {
        console.log("No data available yet or empty data returned");
        return;
      }

      // Process the data - convert to proper types
      const processedData: ConnectionEntry[] = data.map(entry => ({
        timestamp: new Date(entry.timestamp),
        status: entry.status as 'UP' | 'DOWN' | 'UNKNOWN',
        latency: entry.status === 'DOWN' ? 0 : parseInt(entry.latency || '0'),
        uptime: entry.uptime || '0s',
        downtime: entry.downtime || '0s',
        total_changes: parseInt(entry.total_changes || '0'),
        message: entry.message || '',
        isDown: entry.status === 'DOWN'
      }));

      setConnectionData(processedData);
      setTotalEvents(processedData.length);
      setLastUpdated(new Date().toLocaleTimeString());
      console.log("Data fetched successfully:", processedData);
    } catch (error) {
      console.error("Error fetching connection data:", error);
      // Don't update state on error to keep existing data
    }
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
  }, [connectionData, statusStartTime]);

  // Fetch data on initial load
  useEffect(() => {
    const fetchDataWrapper = () => {
      fetchData().catch(console.error);
    };

    fetchDataWrapper();

    // Fetch data every 30 seconds
    const interval = setInterval(fetchDataWrapper, 30000);
    return () => clearInterval(interval);
  }, []);

  const statusColor: Record<string, string> = {
    UP: "green",
    DOWN: "red",
    UNKNOWN: "gray"
  };

  const currentStatus = connectionData.length > 0
      ? connectionData[connectionData.length - 1].status
      : "UNKNOWN";
      
  // Get current events for pagination
  const indexOfLastEvent = currentPage * eventsPerPage;
  const indexOfFirstEvent = indexOfLastEvent - eventsPerPage;
  const currentEvents = connectionData.slice().reverse().slice(indexOfFirstEvent, indexOfLastEvent);
  
  // Calculate total pages
  const totalPages = Math.ceil(connectionData.length / eventsPerPage);
  
  // Change page
  const paginate = (pageNumber: number) => setCurrentPage(pageNumber);
  
  // Generate page numbers for pagination
  const pageNumbers = [];
  for (let i = 1; i <= totalPages; i++) {
    pageNumbers.push(i);
  }

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
                    tickFormatter={(timestamp: Date) => timestamp.toLocaleTimeString()}
                />
                <YAxis />
                <Tooltip
                    labelFormatter={(timestamp: Date) => timestamp.toLocaleTimeString()}
                    formatter={(value: number) => [`${value} ms`, 'Latency']}
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
          <h2 style={{ fontSize: '18px', fontWeight: 'bold', marginBottom: '10px' }}>
            Connection Events 
            <span style={{ fontWeight: 'normal', fontSize: '14px', marginLeft: '10px' }}>
              (Showing {indexOfFirstEvent + 1}-{Math.min(indexOfLastEvent, totalEvents)} of {totalEvents})
            </span>
          </h2>
          
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
            {currentEvents.map((entry, index) => (
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
          
          {/* Pagination */}
          {totalPages > 1 && (
            <div style={{ 
              display: 'flex', 
              justifyContent: 'center', 
              marginTop: '20px',
              gap: '8px'
            }}>
              <button 
                onClick={() => paginate(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1}
                style={{
                  padding: '5px 10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  backgroundColor: currentPage === 1 ? '#f3f4f6' : 'white',
                  cursor: currentPage === 1 ? 'not-allowed' : 'pointer'
                }}
              >
                Previous
              </button>
              
              {pageNumbers.map(number => {
                // Show limited page numbers with ellipsis
                if (
                  number === 1 || 
                  number === totalPages || 
                  (number >= currentPage - 1 && number <= currentPage + 1)
                ) {
                  return (
                    <button
                      key={number}
                      onClick={() => paginate(number)}
                      style={{
                        padding: '5px 10px',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                        backgroundColor: currentPage === number ? '#3B82F6' : 'white',
                        color: currentPage === number ? 'white' : 'black',
                        cursor: 'pointer'
                      }}
                    >
                      {number}
                    </button>
                  );
                } else if (
                  (number === 2 && currentPage > 3) ||
                  (number === totalPages - 1 && currentPage < totalPages - 2)
                ) {
                  return <span key={number} style={{ alignSelf: 'center' }}>...</span>;
                }
                return null;
              })}
              
              <button 
                onClick={() => paginate(Math.min(totalPages, currentPage + 1))}
                disabled={currentPage === totalPages}
                style={{
                  padding: '5px 10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  backgroundColor: currentPage === totalPages ? '#f3f4f6' : 'white',
                  cursor: currentPage === totalPages ? 'not-allowed' : 'pointer'
                }}
              >
                Next
              </button>
            </div>
          )}
        </div>
      </div>
  );
};

export default SimpleMonitor;