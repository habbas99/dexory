import React from 'react';

const Summary = ({ summary, onSummaryCountClick }) => {
  return (
    <div style={{ border: '1px solid #ddd', padding: '15px', borderRadius: '5px' }}>
      <ul>
        {Object.keys(summary).map((key, index) => (
          <li key={index}>
            {key}: 
            <span
              onClick={() => onSummaryCountClick(key)}
              style={{ cursor: 'pointer', color: '#007bff', marginLeft: '10px' }}
            >
              {summary[key]}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Summary;
