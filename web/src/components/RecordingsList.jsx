// web/src/components/RecordingsList.jsx
import React from 'react';
import { AlertTriangle, Play, Pause, Download, Edit, Lock, Unlock, Trash2, FileAudio } from 'lucide-react'; // Added FileAudio

// Helper to format file size (optional)
const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

// Simple presentational component
const RecordingsList = ({
    recordings,     // Array of recording objects (now matching GORM model)
    onPlayPause,    // Function to handle play/pause
    playingFile,    // The URL/path of the currently playing file
    isAudioPlaying, // Boolean indicating if audio is currently playing
    onEdit,         // Function to open edit modal
    onProtect,      // Function to toggle protection
    onDelete        // Function to open delete modal
}) => {
    // Loading State
    if (recordings === null) {
        return (
            <div className="bg-gray-800 p-8 rounded-lg shadow-md text-center text-gray-400 border border-gray-700">
                <FileAudio className="w-12 h-12 mx-auto mb-4 animate-pulse" />
                <p>Loading recordings...</p>
            </div>
        );
    }

    // Empty State
    if (recordings.length === 0) {
        return (
             <div className="bg-gray-800 p-8 rounded-lg shadow-md text-center text-gray-500 border border-gray-700">
                <AlertTriangle className="w-12 h-12 mx-auto mb-4" />
                <p>No recordings found.</p>
                <p className="text-xs mt-2">Start recording manually or enable Auto-Record.</p>
            </div>
        );
    }

    // Recordings Table
    return (
        <section>
            <h2 className="text-xl font-semibold border-b-2 border-gray-700 pb-2 mb-4 text-gray-100">Recordings</h2>
            <div className="bg-gray-800 rounded-lg shadow-md overflow-hidden border border-gray-700">
                <div className="overflow-x-auto"> {/* Enable horizontal scroll on small screens */}
                    <table className="w-full text-left text-sm min-w-[600px]"> {/* Min width for table */}
                        <thead className="bg-gray-700/50 text-gray-300 uppercase text-xs">
                            <tr>
                                <th className="p-3 font-semibold">Name / Filename</th>
                                <th className="p-3 font-semibold">Date Created</th>
                                {/* <th className="p-3 font-semibold">Size</th> Optional */}
                                <th className="p-3 font-semibold text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-700">
                            {recordings.map(rec => {
                                const isPlaying = playingFile === `/recordings/${rec.Filename}` && isAudioPlaying; // Check GORM Filename
                                const isSelected = playingFile === `/recordings/${rec.Filename}`; // Just selected, not necessarily playing

                                return (
                                    <tr key={rec.ID} className={`hover:bg-gray-700/50 transition-colors ${isSelected ? 'bg-blue-900/30' : ''}`}>
                                        {/* Name / Filename */}
                                        <td className="p-3 font-medium text-gray-100 align-middle">
                                            <div>{rec.Name || rec.Filename}</div>
                                            {rec.Name && <div className="text-xs text-gray-500 font-mono">{rec.Filename}</div>}
                                        </td>
                                        {/* Date */}
                                        <td className="p-3 text-gray-400 align-middle whitespace-nowrap">
                                             {/* Format date nicely */}
                                             {new Date(rec.CreatedAt).toLocaleString(undefined, {
                                                 year: 'numeric', month: 'short', day: 'numeric',
                                                 hour: 'numeric', minute: '2-digit'
                                             })}
                                        </td>
                                        {/* Actions */}
                                        <td className="p-3 align-middle">
                                            <div className="flex items-center justify-center space-x-1 sm:space-x-2">
                                                <button onClick={() => onPlayPause(rec.Filename)} className={`p-2 rounded-md transition-colors ${isPlaying ? 'text-blue-400 bg-gray-700' : 'text-gray-400 hover:bg-gray-700 hover:text-white'}`} title={isPlaying ? 'Pause' : 'Play'}>
                                                    {isPlaying ? <Pause size={18} /> : <Play size={18} />}
                                                </button>
                                                <a href={`/recordings/${rec.Filename}`} download={rec.Name || rec.Filename} className="p-2 text-gray-400 hover:bg-gray-700 hover:text-white rounded-md transition-colors" title="Download">
                                                    <Download size={18} />
                                                </a>
                                                <button onClick={() => onEdit(rec)} className="p-2 text-gray-400 hover:bg-gray-700 hover:text-white rounded-md transition-colors" title="Edit Metadata">
                                                    <Edit size={18} />
                                                </button>
                                                <button onClick={() => onProtect(rec)} className={`p-2 rounded-md transition-colors ${rec.Protected ? 'text-yellow-400 hover:bg-gray-700' : 'text-gray-400 hover:bg-gray-700 hover:text-white'}`} title={rec.Protected ? 'Unlock File' : 'Protect File'}>
                                                    {rec.Protected ? <Lock size={18} /> : <Unlock size={18}/>}
                                                </button>
                                                <button onClick={() => !rec.Protected && onDelete(rec)} disabled={rec.Protected} className="p-2 text-red-500 hover:text-red-400 hover:bg-red-900/30 disabled:text-gray-600 disabled:cursor-not-allowed rounded-md transition-colors" title={rec.Protected ? 'File is Protected' : 'Delete File'}>
                                                    <Trash2 size={18} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                 </div>
            </div>
        </section>
    );
};

export default RecordingsList;
