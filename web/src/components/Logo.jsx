// web/src/components/Logo.jsx
import React from 'react';

// Recolorable SVG using currentColor, matching the agreed-upon reel-to-reel design.
const Logo = ({ className }) => (
  <svg 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg" 
    stroke="currentColor" 
    strokeWidth="1.5" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    className={className}
  >
    {/* Main body of the recorder */}
    <rect x="2" y="4" width="20" height="16" rx="2"/>
  
    {/* Left Reel */}
    <circle cx="7" cy="10" r="4"/>
    <circle cx="7" cy="10" r="1" fill="currentColor"/>
    {/* Spokes for left reel */}
    <path d="M7 10 l 2.82 2.82"/>
    <path d="M7 10 l -2.82 -2.82"/>
    <path d="M7 10 l 2.82 -2.82"/>
    <path d="M7 10 l -2.82 2.82"/>
  
    {/* Right Reel */}
    <circle cx="17" cy="10" r="4"/>
    <circle cx="17" cy="10" r="1" fill="currentColor"/>
    {/* Spokes for right reel */}
    <path d="M17 10 l 2.82 2.82"/>
    <path d="M17 10 l -2.82 -2.82"/>
    <path d="M17 10 l 2.82 -2.82"/>
    <path d="M17 10 l -2.82 2.82"/>
    
    {/* Tape Path */}
    <path d="M7 6 C 10 6, 14 6, 17 6"/>
    <path d="M7 14 C 9 14, 9 16, 11 16"/>
    <path d="M17 14 C 15 14, 15 16, 13 16"/>
    <circle cx="10" cy="13" r="0.5" fill="currentColor"/>
    <circle cx="14" cy="13" r="0.5" fill="currentColor"/>
  
    {/* Control Knobs */}
    <circle cx="6" cy="17" r="1"/>
    <circle cx="9" cy="17" r="1"/>
  </svg>
);

export default Logo;
