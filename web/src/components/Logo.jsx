// web/src/components/Logo.jsx
import React from 'react';

// Recolorable SVG using currentColor
const Logo = ({ className }) => (
 <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    // Use currentColor for stroke, allows Tailwind text color classes to work
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    className={`lucide lucide-disc-3 ${className}`} // Apply passed className
  >
    {/* Outer circle */}
    <circle cx="12" cy="12" r="10" />
    {/* Inner hub - stroke only */}
    <circle cx="12" cy="12" r="2" />
     {/* Center dot - fill with currentColor */}
    <path d="M12 14a2 2 0 1 0 0-4 2 2 0 0 0 0 4Z" fill="currentColor"/>
    {/* Cutouts/spokes */}
    <path d="M12 2v4" />
    <path d="m16.95 3.05 2.12 2.12" />
    <path d="M22 12h-4" />
    <path d="m19.07 16.95-2.12 2.12" />
    <path d="M12 22v-4" />
    <path d="m7.05 19.07-2.12-2.12" />
    <path d="M2 12h4" />
    <path d="m4.93 7.05 2.12-2.12" />
 </svg>
);

export default Logo;

