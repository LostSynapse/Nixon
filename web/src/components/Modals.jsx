import React from 'react';

// A reusable modal backdrop and container
const Modal = ({ children, onClose, title }) => (
  <div className="fixed inset-0 bg-black bg-opacity-75 flex justify-center items-center z-50">
    <div className="bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4">
      <div className="p-4 border-b border-gray-700 flex justify-between items-center">
        <h2 className="text-xl font-semibold">{title}</h2>
        <button onClick={onClose} className="text-gray-400 hover:text-white">&times;</button>
      </div>
      <div className="p-4">
        {children}
      </div>
    </div>
  </div>
);

// Settings Modal Component
export const SettingsModal = ({ config, onConfigChange, onSave, onClose }) => {
  // Placeholder content for settings
  return (
    <Modal title="Settings" onClose={onClose}>
      <p className="text-gray-400">Settings will be configured here.</p>
      {/* Example of a setting */}
      <div className="mt-4">
        <label className="block text-sm font-medium text-gray-300">Listen Address</label>
        <input
          type="text"
          value={config?.web?.listenAddress || ''}
          readOnly
          className="mt-1 block w-full bg-gray-700 border-gray-600 rounded-md shadow-sm text-gray-300"
        />
      </div>
       <div className="mt-6 flex justify-end">
        <button onClick={onClose} className="bg-gray-600 hover:bg-gray-500 text-white font-bold py-2 px-4 rounded">
          Close
        </button>
      </div>
    </Modal>
  );
};

// Audio Device Modal Component
export const AudioDeviceModal = ({ audioCaps, onClose }) => {
  // Placeholder content for audio device selection
  const devices = audioCaps?.devices || [];
  
  return (
    <Modal title="Audio Devices" onClose={onClose}>
      {devices.length > 0 ? (
        <ul className="space-y-2">
          {devices.map((device, index) => (
            <li key={index} className="bg-gray-700 p-3 rounded-md">
              <p className="font-semibold">{device.description}</p>
              <p className="text-sm text-gray-400">Driver: {device.driver}</p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-gray-400">No audio devices found or an error occurred.</p>
      )}
       <div className="mt-6 flex justify-end">
        <button onClick={onClose} className="bg-gray-600 hover:bg-gray-500 text-white font-bold py-2 px-4 rounded">
          Close
        </button>
      </div>
    </Modal>
  );
};
