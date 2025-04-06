// src/components/SimulationModal.js
import React from 'react';
import { Modal, Button } from 'react-bootstrap';
import { Chart as ChartJS } from 'chart.js/auto';
import { Line } from 'react-chartjs-2';

const SimulationModal = ({ show, handleClose, simulationData }) => {
    // Destructure simulationData for ease of use
    const withDCData = simulationData.with_data_centers || [];
    const withoutDCData = simulationData.without_data_centers || [];

    // Assuming both series cover the same years
    const labels = withDCData.map(point => point.year);
    const withDCSeries = withDCData.map(point => point.total_temperature);
    const withoutDCSeries = withoutDCData.map(point => point.total_temperature);

    // Compute difference and apply a scaling factor (e.g., multiply by 10)
    const scalingFactor = 10;
    const diffSeries = withDCSeries.map((temp, i) =>
        parseFloat(((temp - withoutDCSeries[i]) * scalingFactor).toFixed(2))
    );

    const chartData = {
        labels,
        datasets: [
            {
                // Bar dataset for (scaled) temperature difference
                type: 'bar',
                label: `Temperature Difference x${scalingFactor} (°C)`,
                data: diffSeries,
                backgroundColor: 'rgba(153,102,255,0.5)',
                borderWidth: 1,
                order: 0,
                yAxisID: 'yDifference' // Assign to separate axis
            },
            {
                // Line dataset: with data centers
                type: 'line',
                label: 'Total Temperature (with DC)',
                data: withDCSeries,
                borderColor: 'rgba(75,192,192,1)',
                backgroundColor: 'rgba(75,192,192,0.2)',
                fill: false,
                tension: 0.2,
                pointRadius: 3,
                pointHoverRadius: 5,
                order: 1,
                yAxisID: 'yTemp' // Assign to main temperature axis
            },
            {
                // Line dataset: without data centers
                type: 'line',
                label: 'Total Temperature (without DC)',
                data: withoutDCSeries,
                borderColor: 'rgba(255,99,132,1)',
                backgroundColor: 'rgba(255,99,132,0.2)',
                fill: false,
                tension: 0.2,
                pointRadius: 3,
                pointHoverRadius: 5,
                order: 1,
                yAxisID: 'yTemp' // Assign to main temperature axis
            }
        ]
    };

    const chartOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: true,
                position: 'top'
            },
            title: {
                display: true,
                text: 'Climate Simulation Over Time',
                font: {
                    size: 18
                }
            }
        },
        scales: {
            yTemp: {
                type: 'linear',
                position: 'left',
                title: {
                    display: true,
                    text: 'Temperature (°C)',
                    font: {
                        size: 14
                    }
                },
                ticks: {
                    font: {
                        size: 12
                    }
                }
            },
            yDifference: {
                type: 'linear',
                position: 'right',
                title: {
                    display: true,
                    text: `Temperature Difference x${scalingFactor} (°C)`,
                    font: {
                        size: 14
                    }
                },
                ticks: {
                    font: {
                        size: 12
                    }
                },
                grid: {
                    drawOnChartArea: false // Remove overlapping grid lines
                }
            },
            x: {
                title: {
                    display: true,
                    text: 'Year',
                    font: {
                        size: 14
                    }
                },
                ticks: {
                    font: {
                        size: 12
                    }
                }
            }
        }
    };

    return (
        <Modal show={show} onHide={handleClose} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>Simulation Results</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <div style={{ width: '100%', height: '400px' }}>
                    <Line data={chartData} options={chartOptions} />
                </div>
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
