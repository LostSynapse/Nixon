// web/src/components/DiskUsage.jsx
import React from 'react';
import { HardDrive } from 'lucide-react';

// Simple presentational component
const DiskUsage = ({ usage }) => {
    // Ensure usage is a number between 0 and 100
    const percentage = Math.max(0, Math.min(100, usage || 0));

    // Determine color based on usage percentage
    let colorClass = 'bg-green-500'; // Default green
    if (percentage > 90) { // Critical Red
      colorClass = 'bg-red-500';
    } else if (percentage > 75) { // Warning Yellow
      colorClass = 'bg-yellow-500';
    }

    return (
        <footer className="fixed bottom-0 left-0 right-0 bg-gray-800/90 backdrop-blur-sm p-3 shadow-inner z-10 border-t border-gray-700">
            <div className="flex items-center justify-center max-w-4xl mx-auto space-x-3">
                <HardDrive className="w-5 h-5 text-gray-400 flex-shrink-0" aria-hidden="true" />
                {/* Progress Bar Container */}
                <div className="w-full bg-gray-600 rounded-full h-4 overflow-hidden" title={`Disk Usage: ${percentage}%`}>
                    {/* Progress Bar Fill */}
                    <div
                      className={`h-4 rounded-full transition-all duration-300 ease-out ${colorClass}`}
                      style={{ width: `${percentage}%` }}
                      role="progressbar"
                      aria-valuenow={percentage}
                      aria-valuemin="0"
                      aria-valuemax="100"
                    ></div>
                </div>
                {/* Percentage Text */}
                <span className="text-sm font-semibold text-white w-12 text-right flex-shrink-0">
                    {percentage}%
                </span>
            </div>
        </footer>
    );
};

export default DiskUsage;
