// src/components/ReportList.js
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Container, Row, Col } from 'react-bootstrap';
import { renderStatusBadge, renderDateStr } from './utils';
import CreateReportModal from './CreateReportModal';

const ReportList = () => {
  const [reports, setReports] = useState([]);
  const [showModal, setShowModal] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetchReports();
  }, []);

  const fetchReports = async () => {
    try {
      const response = await axios.get('/inventory-comparison-reports');
      setReports(response.data);
    } catch (error) {
      console.error('Error fetching reports:', error);
    }
  };

  const handleCreateReport = async ({ bulkScanFileName, csvFile }) => {
    const formData = new FormData();
    formData.append('bulkScanFileName', bulkScanFileName);
    formData.append('csvFile', csvFile);

    try {
      await axios.post('/inventory-comparison-reports', formData);
      fetchReports();
      handleCloseModal();
    } catch (error) {
      console.error('Error creating report:', error);
    }
  };

  const handleRowClick = (reportId) => {
    navigate(`/report/${reportId}`);
  };

  const handleOpenModal = () => setShowModal(true);
  const handleCloseModal = () => setShowModal(false);

  return (
    <Container>
      <Row className="my-4">
        <Col className="text-end">
          <Button variant="primary" onClick={handleOpenModal}>
            Create Report
          </Button>
        </Col>
      </Row>

      <Row className="my-4">
        <Col>
          <Table striped bordered hover>
            <thead>
              <tr>
                <th>ID</th>
                <th>Bulk Scan File</th>
                <th>Reference File</th>
                <th>Created</th>
                <th>Updated</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {reports.map((report) => (
                <tr key={report.id} onClick={() => handleRowClick(report.id)} style={{ cursor: 'pointer' }}>
                  <td>{report.id}</td>
                  <td>{report.bulkScanFileName}</td>
                  <td>{report.referenceFileName}</td>
                  <td>{renderDateStr(report.createdAt)}</td>
                  <td>{renderDateStr(report.updatedAt)}</td>
                  <td>{renderStatusBadge(report.status)}</td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Col>
      </Row>

      <CreateReportModal
        showModal={showModal}
        handleCloseModal={handleCloseModal}
        handleCreateReport={handleCreateReport}
      />
    </Container>
  );
};

export default ReportList;
