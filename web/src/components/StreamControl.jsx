// web/src/components/StreamControl.jsx
import React from 'react';
import { Power, Wifi, Users, ArrowUp } from 'lucide-react'; // Changed icons slightly

// Simple presentational component
const StreamControl = ({ title, status, onToggle, listeners, peak }) => (
    <div className="bg-gray-800 p-4 rounded-lg shadow-md flex flex-col items-center justify-between border border-gray-700">
      <div className="flex items-center justify-between w-full mb-3">
        {/* Header with Icon and Title */}
        <div className="flex items-center">
            {title.includes('SRT') ? <Wifi size={20} className="mr-2 text-cyan-400" /> : <Wifi size={20} className="mr-2 text-orange-400" />} {/* Use Wifi for streams */}
            <h3 className="text-lg font-semibold text-gray-100">{title}</h3>
        </div>
        {/* Listener count for Icecast */}
        {title.includes('Icecast') && status && (
            <div className="flex items-center text-xs text-gray-400" title="Current / Peak Listeners">
                <Users size={14} className="mr-1"/> {listeners ?? 0}
                <ArrowUp size={14} className="ml-2 mr-1 text-gray-500"/> {peak ?? 0}
            </div>
        )}
      </div>
      {/* Status Indicator */}
      <div className={`text-sm font-bold px-3 py-1 rounded-full mb-4 inline-block ${status ? 'bg-red-600 text-white animate-pulse' : 'bg-gray-600 text-gray-300'}`}>
          {status ? 'LIVE' : 'OFFLINE'}
      </div>
      {/* Action Button */}
      <button
        onClick={onToggle}
        className={`w-full flex items-center justify-center p-3 rounded-lg font-bold text-white transition-colors duration-150 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-800 ${
            status
             ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500'
             : 'bg-green-600 hover:bg-green-700 focus:ring-green-500'
        }`}
      >
        <Power className="w-5 h-5 mr-2" />
        {status ? 'Stop Stream' : 'Go Live'}
      </button>
    </div>
);

export default StreamControl;

