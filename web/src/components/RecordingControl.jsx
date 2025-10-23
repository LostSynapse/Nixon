// web/src/components/RecordingControl.jsx
import React from 'react';
import { Mic, Square, CircleDot, Divide } from 'lucide-react'; // Updated icons

// Simple presentational component
const RecordingControl = ({ status, currentFile, onToggle, onSplit, onAutoToggle, autoRecordEnabled }) => (
    <div className="bg-gray-800 p-4 rounded-lg shadow-md col-span-1 md:col-span-2 border border-gray-700">
        {/* Header with Auto-Toggle */}
        <div className="flex justify-between items-center mb-3">
            <div className="flex items-center">
                <Mic className="w-6 h-6 mr-2 text-red-400" />
                <h3 className="text-lg font-semibold text-gray-100">Recording</h3>
            </div>
            {/* Improved Auto-Record Toggle */}
            <label htmlFor="auto-record-toggle" className="flex items-center space-x-2 cursor-pointer" title="Toggle Automatic Recording (VAD)">
                <span className="text-xs text-gray-400">Auto</span>
                {/* B1: Use form-switch if available or style manually */}
                <button
                    id="auto-record-toggle"
                    type="button"
                    onClick={onAutoToggle}
                    className={`relative inline-flex items-center h-5 w-10 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-800 ${autoRecordEnabled ? 'bg-blue-600' : 'bg-gray-600'}`}
                    role="switch"
                    aria-checked={autoRecordEnabled}
                 >
                    <span
                        aria-hidden="true"
                        className={`inline-block h-3 w-3 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${autoRecordEnabled ? 'translate-x-5' : 'translate-x-1'}`}
                    />
                </button>
            </label>
        </div>
        {/* Status Indicator */}
        <div className={`text-sm font-bold px-3 py-1 rounded-full mb-4 inline-flex items-center ${status ? 'bg-red-600 text-white animate-pulse' : 'bg-gray-600 text-gray-300'}`}>
          {status ? <CircleDot size={14} className="mr-1.5"/> : <Square size={14} className="mr-1.5"/>}
          {status ? 'RECORDING' : 'IDLE'}
        </div>
        {/* Current File Display */}
        {status && currentFile && (
            <p className="text-xs text-gray-400 mb-4 break-all" title={currentFile}>
                File: <span className="font-mono">{currentFile}</span>
            </p>
        )}
        {/* Action Buttons */}
        <div className="grid grid-cols-2 gap-3">
            <button
                onClick={onToggle} // onToggle now likely comes from useNixonApi
                className={`w-full flex items-center justify-center p-3 rounded-lg font-bold text-white transition-colors duration-150 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-800 ${
                    status
                     ? 'bg-red-600 hover:bg-red-700 focus:ring-red-500'
                     : 'bg-blue-600 hover:bg-blue-700 focus:ring-blue-500'
                }`}
             >
                {status ? <Square size={18} className="mr-2"/> : <CircleDot size={18} className="mr-2"/>}
                {status ? 'Stop' : 'Start'}
            </button>
            <button
                onClick={onSplit}
                disabled={!status} // Disable split if not recording
                className="w-full flex items-center justify-center p-3 rounded-lg font-bold text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 focus:ring-offset-gray-800 disabled:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
                title="Split Recording (Starts New File)"
            >
               <Divide size={18} className="mr-2"/> Split
            </button>
        </div>
    </div>
);

export default RecordingControl;
