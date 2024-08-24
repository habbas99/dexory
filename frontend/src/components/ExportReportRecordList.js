// ExportedFilesList.js

import React from 'react';
import { ListGroup, Button } from 'react-bootstrap';
import {renderStatusBadge} from "./utils";

const ExportReportRecordList = ({ exportReportRecords }) => {
    const handleDownload = async (exportReportRecordId) => {
        try {
            const response = await fetch(`/export-report-records/${exportReportRecordId}/download`);
            if (!response.ok) {
                throw new Error('failed to download file');
            }

            // extract the filename from the 'Content-Disposition' header
            const contentDisposition = response.headers.get('Content-Disposition');
            console.log("contentDisposition = " + contentDisposition);

            // updated regex to match filenames with or without quotes
            const filenameMatch = contentDisposition ? contentDisposition.match(/filename="?([^"]+)"?/) : null;
            console.log("filenameMatch = ", filenameMatch);

            // use the matched filename or fallback to a default name
            const filename = filenameMatch && filenameMatch[1];
            console.log("filename = " + filename);

            // convert response to a Blob
            const blob = await response.blob();

            // create a temporary URL and trigger the download
            const downloadUrl = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(downloadUrl);
        } catch (error) {
            console.error('Download error:', error);
        }
    };

    return (
        <div>
            <ListGroup>
                {exportReportRecords.length > 0 ? (
                    exportReportRecords.map((exportReportRecord) => (
                        <ListGroup.Item key={exportReportRecord.id}>
                            <div className="d-flex justify-content-between align-items-center">
                                <span>{exportReportRecord.fileName}</span>
                                {exportReportRecord.status === 'completed' ? (
                                    <Button
                                        variant="link"
                                        onClick={() => handleDownload(exportReportRecord.id)}
                                    >
                                        Download
                                    </Button>
                                ) : (
                                    <span>{renderStatusBadge(exportReportRecord.status)}</span>
                                )}
                            </div>
                        </ListGroup.Item>
                    ))
                ) : (
                    <p>No exported files found.</p>
                )}
            </ListGroup>
        </div>
    );
};

export default ExportReportRecordList;
