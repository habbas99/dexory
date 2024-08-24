import React, { useState } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';
import axios from 'axios';

const ExportReportModal = ({ show, handleClose, reportId }) => {
    const [exportType, setExportType] = useState('json');

    const handleExport = async () => {
        try {
            const response = await axios.post('/export-report-records', {
                reportRecordId: Number(reportId),
                reportType: exportType,
            });
            console.log('Export successful:', response.data);
            handleClose(); // Close the modal after export
        } catch (error) {
            console.error('Error exporting report:', error);
        }
    };

    return (
        <Modal show={show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>Export Report</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group controlId="exportType">
                        <Form.Label>Select Export Type</Form.Label>
                        <Form.Control
                            as="select"
                            value={exportType}
                            onChange={(e) => setExportType(e.target.value)}
                        >
                            <option value="json">JSON</option>
                            <option value="csv">CSV</option>
                        </Form.Control>
                    </Form.Group>
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={handleClose}>Close</Button>
                <Button variant="primary" onClick={handleExport}>Export</Button>
            </Modal.Footer>
        </Modal>
    );
};

export default ExportReportModal;
