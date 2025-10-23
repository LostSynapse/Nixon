// web/src/components/Logo.jsx
import React from 'react';

const Logo = ({ className }) => {
  return (
    <svg
      className={className}
      width="100"
      height="100"
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="25" cy="50" r="15" stroke="currentColor" strokeWidth="4" />
      <circle cx="75" cy="50" r="15" stroke="currentColor" strokeWidth="4" />
      <path
        d="M25 35L75 35"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <path
        d="M25 65L75 65"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <path
        d="M40 50L25 65"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <path
        d="M60 50L75 35"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <circle cx="25" cy="50" r="3" fill="currentColor" />
      <circle cx="75" cy="50" r="3" fill="currentColor" />
    </svg>
  );
};

export default Logo;
