// src/components/SimulationModal.js
import React from 'react';
import { Modal, Button } from 'react-bootstrap';
import { Line } from 'react-chartjs-2';
import 'chart.js/auto';

const SimulationModal = ({ show, handleClose, simulationData }) => {
    // Prepare data for the graph.
    const labels = simulationData.data.map(point => point.Year);
    const temperatureData = simulationData.data.map(point => point.TotalTemperature);

    const chartData = {
        labels,
        datasets: [
            {
                label: 'Total Temperature (°C)',
                data: temperatureData,
                borderColor: 'rgba(75,192,192,1)',
                backgroundColor: 'rgba(75,192,192,0.2)',
                fill: true,
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
                <Line data={chartData} options={{ responsive: true, maintainAspectRatio: false }} style={{ height: '300px' }} />
                <div className="mt-3">
                    <p><strong>Total Time to Uninhabitability:</strong> {simulationData.total_time_to_end} years</p>
                    <p><strong>Data Center Contribution Shortens Time by:</strong> {simulationData.time_datacenters_removed} years</p>
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
