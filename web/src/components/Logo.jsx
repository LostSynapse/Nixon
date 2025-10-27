import React from 'react';

// This component now wraps the SVG and adds the status indicator logic.
const Logo = ({ isConnected, appState = '' }) => {
  const getStatusColor = () => {
    if (!isConnected) return 'bg-yellow-500'; // Connecting or disconnected
    if (appState.includes('recording')) return 'bg-red-500'; 
    return 'bg-green-500'; // Connected and idle
  };

  return (
    <div className="relative h-16 w-16">
      <svg 
        viewBox="0 0 24 24" 
        fill="none" 
        xmlns="http://www.w3.org/2000/svg" 
        stroke="currentColor" 
        strokeWidth="1.5" 
        strokeLinecap="round" 
        strokeLinejoin="round"
        className="h-full w-full"
      >
        {/* SVG content from your file remains here */}
        <rect x="2" y="4" width="20" height="16" rx="2"/>
        <circle cx="7" cy="10" r="4"/><circle cx="7" cy="10" r="1" fill="currentColor"/>
        <path d="M7 10 l 2.82 2.82"/><path d="M7 10 l -2.82 -2.82"/><path d="M7 10 l 2.82 -2.82"/><path d="M7 10 l -2.82 2.82"/>
        <circle cx="17" cy="10" r="4"/><circle cx="17" cy="10" r="1" fill="currentColor"/>
        <path d="M17 10 l 2.82 2.82"/><path d="M17 10 l -2.82 -2.82"/><path d="M17 10 l 2.82 -2.82"/><path d="M17 10 l -2.82 2.82"/>
        <path d="M7 6 C 10 6, 14 6, 17 6"/><path d="M7 14 C 9 14, 9 16, 11 16"/><path d="M17 14 C 15 14, 15 16, 13 16"/>
        <circle cx="10" cy="13" r="0.5" fill="currentColor"/><circle cx="14" cy="13" r="0.5" fill="currentColor"/>
        <circle cx="6" cy="17" r="1"/><circle cx="9" cy="17" r="1"/>
      </svg>
      <span
        className={`absolute bottom-1 right-1 block h-4 w-4 rounded-full ${getStatusColor()} ring-2 ring-gray-900`}
        title={isConnected ? `Status: ${appState}` : 'Connecting...'}
      />
    </div>
  );
};

export default Logo;
