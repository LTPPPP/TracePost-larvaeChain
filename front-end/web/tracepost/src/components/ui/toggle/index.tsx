'use client';
import React from 'react';

interface ToggleButtonProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
}

const ToggleButton: React.FC<ToggleButtonProps> = ({ checked, onChange }) => {
  return (
    <label className='relative inline-flex items-center cursor-pointer'>
      <input type='checkbox' className='sr-only peer' checked={checked} onChange={(e) => onChange(e.target.checked)} />
      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
      <span className='ml-2 text-sm font-medium text-gray-900 dark:text-gray-300'>{checked ? 'On' : 'Blocked'}</span>
    </label>
  );
};

export default ToggleButton;
