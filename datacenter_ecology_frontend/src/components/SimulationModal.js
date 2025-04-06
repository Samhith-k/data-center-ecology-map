// src/components/SimulationModal.js
import React from 'react';
import { Modal, Button } from 'react-bootstrap';
import { Line } from 'react-chartjs-2';
import 'chart.js/auto';

const SimulationModal = ({ show, handleClose, simulationData }) => {
    // Destructure simulationData for ease of use.
    // We now expect simulationData to have: 
    //   with_data_centers, without_data_centers, total_time_to_end, time_datacenters_removed
    const withDCData = simulationData.with_data_centers || [];
    const withoutDCData = simulationData.without_data_centers || [];

    // Assuming both series cover the same years.
    const labels = withDCData.map(point => point.year);

    const withDCSeries = withDCData.map(point => point.total_temperature);
    const withoutDCSeries = withoutDCData.map(point => point.total_temperature);

    const chartData = {
        labels,
        datasets: [
            {
                label: 'Total Temperature (with Data Centers)',
                data: withDCSeries,
                borderColor: 'rgba(75,192,192,1)',
                backgroundColor: 'rgba(75,192,192,0.2)',
                fill: false,
                tension: 0.2
            },
            {
                label: 'Total Temperature (without Data Centers)',
                data: withoutDCSeries,
                borderColor: 'rgba(255,99,132,1)',
                backgroundColor: 'rgba(255,99,132,0.2)',
                fill: false,
                tension: 0.2
            }
        ]
    };

    return (
        <Modal show={show} onHide={handleClose} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>Simulation Results</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Line
                    data={chartData}
                    options={{ responsive: true, maintainAspectRatio: false }}
                    style={{ height: '300px' }}
                />
                <div className="mt-3">
                    <p>
                        <strong>Total Time to Uninhabitability (with Data Centers):</strong> {simulationData.total_time_to_end} years
                    </p>
                    <p>
                        <strong>Time Difference (Without Data Centers):</strong> {simulationData.time_datacenters_removed} years
                    </p>
                </div>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={handleClose}>
                    Close
                </Button>
            </Modal.Footer>
        </Modal>
    );
};

export default SimulationModal;
