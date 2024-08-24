import React, { useState, useEffect } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';
import axios from 'axios';

const CreateReportModal = ({ showModal, handleCloseModal, handleCreateReport }) => {
  const [bulkScanFileName, setBulkScanFileName] = useState('');
  const [csvFile, setCsvFile] = useState(null);
  const [bulkScanRecords, setBulkScanRecords] = useState([]);

  useEffect(() => {
    const fetchBulkScanRecords = async () => {
      try {
        const response = await axios.get('/bulk-scan-records');
        setBulkScanRecords(response.data);
        if (response.data.length > 0) {
            setBulkScanFileName(response.data[0].fileName);
        }
      } catch (error) {
        console.error('Error fetching bulk scan records:', error);
      }
    };

    fetchBulkScanRecords();
  }, []);

  const handleFileChange = (e) => {
    setCsvFile(e.target.files[0]);
  };

  const onSubmit = (e) => {
    e.preventDefault();
    console.log("on submit, bulkScanFileName = " + bulkScanFileName)
    handleCreateReport({ bulkScanFileName, csvFile });
    setBulkScanFileName('');
    setCsvFile(null);
  };

  return (
    <Modal show={showModal} onHide={null} backdrop="static">
      <Modal.Header>
        <Modal.Title>Create Comparison Report</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <Form onSubmit={onSubmit}>
          <Form.Group controlId="report">
            <Form.Label>Select Bulk Scan File</Form.Label>
            <Form.Control
              as="select"
              value={bulkScanFileName}
              onChange={(e) => setBulkScanFileName(e.target.value)}
              required
            >
              {bulkScanRecords.map((record) => (
                <option key={record.id} value={record.fileName}>
                  {record.fileName}
                </option>
              ))}
            </Form.Control>
          </Form.Group>

          <Form.Group controlId="formCsvFile" className="my-3">
            <Form.Label>CSV File</Form.Label>
            <Form.Control
              type="file"
              accept=".csv"
              onChange={handleFileChange}
              required
            />
          </Form.Group>
        </Form>
      </Modal.Body>
      <Modal.Footer className="text-end">
        <Button variant="secondary" onClick={handleCloseModal}>
          Close
        </Button>
        <Button variant="success" type="submit" onClick={onSubmit}>
          Create
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default CreateReportModal;
