// web/src/components/StreamControl.jsx
import React from 'react';
import { Radio, Mic, Settings, Cast } from 'lucide-react';

const StreamControl = ({ onSettingsClick }) => {
  // TODO: Add state and handlers for stream buttons
  return (
    <div className="bg-gray-800 p-4 rounded-lg border border-gray-700 flex items-center justify-between">
      <div className="flex items-center space-x-4">
        <button className="p-3 rounded-full bg-red-500 hover:bg-red-600 text-white">
          <Mic size={24} />
        </button>
        <div className="text-sm">
          <p className="font-bold text-white">Recording Status</p>
          <p className="text-gray-400">Not Recording</p>
        </div>
      </div>
      <div className="flex items-center space-x-2">
         <button className="p-2 rounded hover:bg-gray-700 text-gray-400 hover:text-white" title="SRT Stream">
          <Radio size={20} />
        </button>
        <button className="p-2 rounded hover:bg-gray-700 text-gray-400 hover:text-white" title="Icecast Stream">
          <Cast size={20} />
        </button>
        <button onClick={onSettingsClick} className="p-2 rounded hover:bg-gray-700 text-gray-400 hover:text-white" title="Settings">
          <Settings size={20} />
        </button>
      </div>
    </div>
  );
};

export default StreamControl;
