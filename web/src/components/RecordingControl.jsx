import React from 'react';
import { Mic } from 'lucide-react';

const RecordingControl = ({ status, currentFile, onToggle, onSplit, onAutoToggle, autoRecordEnabled }) => (
    <div className="bg-gray-800 p-4 rounded-lg shadow-md col-span-1 md:col-span-2">
        <div className="flex justify-between items-center mb-3">
            <div className="flex items-center"><Mic className="w-6 h-6 mr-2 text-red-400" /><h3 className="text-lg font-semibold">Recording</h3></div>
            <label title="Toggle Automatic Recording" className="flex items-center space-x-2 cursor-pointer">
                <span className="text-xs text-gray-400">Auto</span>
                <div onClick={onAutoToggle} className={`w-10 h-5 flex items-center rounded-full p-1 transition-colors ${autoRecordEnabled ? 'bg-blue-600' : 'bg-gray-600'}`}>
                    <div className={`bg-white w-3 h-3 rounded-full shadow-md transform transition-transform ${autoRecordEnabled ? 'translate-x-5' : ''}`}></div>
                </div>
            </label>
        </div>
      <div className={`text-sm font-bold px-2 py-1 rounded-full mb-4 inline-block ${status ? 'bg-red-500 text-white animate-pulse' : 'bg-gray-600 text-gray-300'}`}>{status ? 'RECORDING' : 'IDLE'}</div>
      {status && <p className="text-xs text-gray-400 mb-4 break-all">File: {currentFile}</p>}
      <div className="grid grid-cols-2 gap-3"><button onClick={() => onToggle(status ? 'stop' : 'start')} className={`w-full flex items-center justify-center p-3 rounded-lg font-bold text-white transition-all ${status ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'}`}>{status ? 'Stop' : 'Start'}</button><button onClick={onSplit} disabled={!status} className="w-full p-3 rounded-lg font-bold text-white bg-purple-600 hover:bg-purple-700 disabled:bg-gray-700 disabled:cursor-not-allowed">Split</button></div>
    </div>
);

export default RecordingControl;

