import React from 'react';
import { Table } from 'react-bootstrap';

const ComparisonTable = ({ data }) => {
    const renderBooleanIcon = (value) => {
        return value ? '✔️' : '❌';
    };

    return (
        <Table striped bordered hover className="mt-3">
            <thead>
                <tr>
                <th>Location</th>
                <th>Scanned</th>
                <th>Occupied</th>
                <th>Actual Barcodes</th>
                <th>Expected Barcodes</th>
                <th>Result</th>
                </tr>
            </thead>
            <tbody>
                {data.map((item, index) => (
                <tr key={index}>
                    <td>{item.location}</td>
                    <td>{renderBooleanIcon(item.scanned)}</td>
                    <td>{renderBooleanIcon(item.occupied)}</td>
                    <td>{item.actualBarcodes.join(', ')}</td>
                    <td>{item.expectedBarcodes.join(', ')}</td>
                    <td>{item.result}</td>
                </tr>
                ))}
            </tbody>
        </Table>
    );
};

export default ComparisonTable;
