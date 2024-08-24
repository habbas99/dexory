import React from 'react';
import { Badge } from 'react-bootstrap';

export const renderStatusBadge = (status) => {
  let variant;
  switch (status) {
    case 'completed':
      variant = 'success';
      break;
    case 'pending':
      variant = 'primary';
      break;
    case 'failed':
      variant = 'danger';
      break;
    default:
      variant = 'secondary';
  }
  return <Badge bg={variant}>{status}</Badge>;
};

export const renderDateStr = (dateStr) => {
  const date = new Date(dateStr);

  const formattedDate = date.toLocaleDateString("en-GB", {
      year: "numeric",
      month: "long",
      day: "numeric",
  });

  const formattedTime = date.toLocaleTimeString("en-GB", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
  });

  return `${formattedDate} ${formattedTime}`; 
}