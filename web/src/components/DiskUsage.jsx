import React from 'react';
import { HardDrive } from 'lucide-react';

const DiskUsage = ({ usage }) => {
    const percentage = usage || 0;
    let colorClass = 'bg-green-500';
    if (percentage > 80) colorClass = 'bg-red-500';
    else if (percentage > 60) colorClass = 'bg-yellow-500';
    return (
        <footer className="fixed bottom-0 left-0 right-0 bg-gray-900/80 backdrop-blur-sm p-3 shadow-inner z-10 border-t border-gray-700"><div className="flex items-center justify-center max-w-4xl mx-auto"><HardDrive className="w-5 h-5 mr-3 text-gray-400"/><div className="w-full bg-gray-700 rounded-full h-4"><div className={`h-4 rounded-full transition-all duration-500 ${colorClass}`} style={{ width: `${percentage}%` }}></div></div><span className="ml-3 text-sm font-semibold text-white w-10 text-center">{percentage}%</span></div></footer>
    );
};

export default DiskUsage;

