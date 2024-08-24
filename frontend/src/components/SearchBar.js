import React from 'react';
import { Form } from 'react-bootstrap';

const SearchComponent = ({ placeholderStr, searchStr, onSearchChange }) => {
  return (
    <Form>
      <Form.Group controlId="search">
        <Form.Control
          type="text"
          placeholder={placeholderStr}
          value={searchStr}
          onChange={(e) => onSearchChange(e.target.value)}
        />
      </Form.Group>
    </Form>
  );
};

export default SearchComponent;
