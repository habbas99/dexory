import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { Container, Row, Col, Form, Button } from 'react-bootstrap';
import { renderStatusBadge, renderDateStr } from './utils';
import Summary from './Summary';
import ComparisonTable from './ComparisonTable';
import SearchBar from './SearchBar';
import ExportReportModal from "./ExportReportModal";
import ExportReportRecordList from "./ExportReportRecordList";

const ReportDetail = () => {
  const { reportId } = useParams();
  const [report, setReport] = useState(null);
  const [comparisonData, setComparisonData] = useState([]);
  const [filteredData, setFilteredData] = useState([]);
  const [searchStr, setSearchStr] = useState('');
  const [summary, setSummary] = useState({});
  const [showExportModal, setShowExportModal] = useState(false);
  const [exportReportRecords, setExportReportRecords] = useState([]);

  useEffect(() => {
    fetchReport();
    fetchComparisonData();
    fetchExportReportRecords();
  }, [reportId]);

  useEffect(() => {
    filterData();
  }, [searchStr, comparisonData]);

  useEffect(() => {
    calculateSummary();
  }, [comparisonData]);

  const fetchReport = async () => {
    try {
      const response = await axios.get(`/inventory-comparison-reports/${reportId}`);
      setReport(response.data);
    } catch (error) {
      console.error('Error fetching report:', error);
    }
  };

  const fetchComparisonData = async () => {
    try {
      const response = await axios.get(`/inventory-comparison-reports/${reportId}/data`);
      setComparisonData(response.data);
      setFilteredData(response.data);
    } catch (error) {
      console.error('Error fetching comparison data:', error);
    }
  };

  const fetchExportReportRecords = async () => {
    try {
      const response = await axios.get(`/inventory-comparison-reports/${reportId}/exports`);
      setExportReportRecords(response.data);
    } catch (error) {
      console.error('Error fetching export report records:', error);
    }
  };

  const filterData = () => {
    const lowerCasedFilter = searchStr.toLowerCase();
    const filtered = comparisonData.filter((item) => {
      return (
        item.location.toLowerCase().includes(lowerCasedFilter) ||
        item.result.toLowerCase().includes(lowerCasedFilter) ||
        item.actualBarcodes.some(
          (barcode) => barcode.toLowerCase().includes(lowerCasedFilter)
        ) ||
        item.expectedBarcodes.some(
          (barcode) => barcode.toLowerCase().includes(lowerCasedFilter)
        )
      );
    });
    setFilteredData(filtered);
  };

  const calculateSummary = () => {
    const summary = {
      "The location was empty, as expected": 0,
      "The location was empty, but it should have been occupied": 0,
      "The location was occupied by the expected items": 0,
      "The location was occupied by the wrong items": 0,
      "The location was occupied, but no barcode could be identified": 0,
      "The location was occupied by an item, but should have been empty": 0,
    };

    comparisonData.forEach((item) => {
      if (summary.hasOwnProperty(item.result)) {
        summary[item.result] += 1;
      }
    });

    setSummary(summary);
  };

  const handleSearchChange = (str) => {
    setSearchStr(str);
  };

  const handleSummaryCountClick = (str) => {
    setSearchStr(str);
  };

  const handleExportModalClose = () => {
    setShowExportModal(false);
    fetchExportReportRecords(); // refresh the export report records
  };

  if (!report) {
    return <div>Loading...</div>;
  }

  return (
    <Container>
      <Row className="my-4">
        <Col>
          <Button variant="secondary" onClick={() => window.history.back()}>Back</Button>
        </Col>
        <Col className="text-end">
          <Button variant="primary" onClick={() => setShowExportModal(true)} disabled={exportReportRecords.length >= 1}
          >Export</Button>
        </Col>
      </Row>

      <Row className="my-4">
        <Col>
          <p><strong>ID:</strong> {report.id}</p>
          <p><strong>Status:</strong> {renderStatusBadge(report.status)}</p>
        </Col>
        <Col>
          <p><strong>Bulk Scan File:</strong> {report.bulkScanFileName}</p>
          <p><strong>Reference File:</strong> {report.referenceFileName}</p>
        </Col>
        <Col>
          <p><strong>Created At:</strong> {renderDateStr(report.createdAt)}</p>
          <p><strong>Updated At:</strong> {renderDateStr(report.updatedAt)}</p>
        </Col>
      </Row>

      <Row className="my-4">
        <Col>
          <h5>Exported Files:</h5>
          <ExportReportRecordList exportReportRecords={exportReportRecords} />
        </Col>
      </Row>

      <Row className="my-4">
        <Col>
          <h5>Summary:</h5>
          <Summary summary={summary} onSummaryCountClick={handleSummaryCountClick} />
        </Col>
      </Row>

      <Row className="my-4">
        <Col>
            <SearchBar placeholderStr="Search by location, barcode, or result" searchStr={searchStr} onSearchChange={handleSearchChange} />
          <ComparisonTable data={filteredData} />
        </Col>
      </Row>

      <ExportReportModal show={showExportModal} handleClose={handleExportModalClose} reportId={reportId} />
    </Container>
  );
};

export default ReportDetail;
