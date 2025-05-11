import React, { useEffect, useState } from 'react';
import { Timeline, Spin, Card, Badge, Typography, Divider, Row, Col, Tag, Alert } from 'antd';
import {
    ClockCircleOutlined,
    EnvironmentOutlined,
    ExclamationCircleOutlined,
    CheckCircleOutlined,
    SwapOutlined,
    ThermometerOutlined,
    AimOutlined,
    WarningOutlined
} from '@ant-design/icons';
import axios from 'axios';
import moment from 'moment';
import { Line } from '@ant-design/charts';
import { MapContainer, TileLayer, Marker, Popup, Polyline } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

const { Title, Text } = Typography;

// Fix Leaflet marker issue
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
    iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
    iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
    shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

const TransactionTimeline = ({ batchId }) => {
    const [loading, setLoading] = useState(true);
    const [timelineData, setTimelineData] = useState(null);
    const [errorMessage, setErrorMessage] = useState('');
    const [activeTab, setActiveTab] = useState('timeline');

    useEffect(() => {
        const fetchTimelineData = async () => {
            try {
                setLoading(true);
                const response = await axios.get(`/api/v1/analytics/timeline/${batchId}`);
                setTimelineData(response.data.data);
                setErrorMessage('');
            } catch (error) {
                console.error('Error fetching timeline data:', error);
                setErrorMessage('Failed to load transaction timeline. Please try again later.');
            } finally {
                setLoading(false);
            }
        };

        if (batchId) {
            fetchTimelineData();
        }
    }, [batchId]);

    if (loading) {
        return (
            <div style={{ textAlign: 'center', padding: '40px' }}>
                <Spin size="large" />
                <p>Loading transaction timeline...</p>
            </div>
        );
    }

    if (errorMessage) {
        return (
            <Alert
                message="Error"
                description={errorMessage}
                type="error"
                showIcon
            />
        );
    }

    if (!timelineData) {
        return (
            <Alert
                message="No Data"
                description="No timeline data available for this batch."
                type="info"
                showIcon
            />
        );
    }

    // Combine all timeline items and sort by timestamp
    const allTimelineItems = [
        ...timelineData.events.map(event => ({ ...event, itemType: 'event' })),
        ...timelineData.transfers.map(transfer => ({ ...transfer, itemType: 'transfer' })),
        ...timelineData.environmentData.map(env => ({ ...env, itemType: 'environment' })),
        ...timelineData.anomalies.map(anomaly => ({ ...anomaly, itemType: 'anomaly' }))
    ].sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

    // Prepare environmental data for charts
    const temperatureData = timelineData.environmentData
        .filter(item => item.temperature !== undefined)
        .map(item => ({
            timestamp: moment(item.timestamp).format('YYYY-MM-DD HH:mm'),
            value: item.temperature,
            type: 'Temperature (°C)'
        }));

    const humidityData = timelineData.environmentData
        .filter(item => item.humidity !== undefined)
        .map(item => ({
            timestamp: moment(item.timestamp).format('YYYY-MM-DD HH:mm'),
            value: item.humidity,
            type: 'Humidity (%)'
        }));

    const environmentChartData = [...temperatureData, ...humidityData];

    // Prepare location data for map
    const locations = allTimelineItems
        .filter(item => item.location && item.location.latitude && item.location.longitude)
        .map(item => ({
            id: item.id,
            type: item.itemType,
            timestamp: item.timestamp,
            position: [item.location.latitude, item.location.longitude],
            description: item.description || (item.itemType === 'transfer' ? `Transfer from ${item.from} to ${item.to}` : ''),
        }));

    // Configuration for environment chart
    const config = {
        data: environmentChartData,
        xField: 'timestamp',
        yField: 'value',
        seriesField: 'type',
        yAxis: {
            title: {
                text: 'Value',
            },
        },
        legend: {
            position: 'top',
        },
        smooth: true,
        animation: {
            appear: {
                animation: 'path-in',
                duration: 1000,
            },
        },
        point: {
            size: 5,
            shape: 'circle',
            style: {
                fill: 'white',
                stroke: '#5B8FF9',
                lineWidth: 2,
            },
        },
    };

    // Map center and bounds
    const mapCenter = locations.length > 0
        ? locations[Math.floor(locations.length / 2)].position
        : [10.762622, 106.660172]; // Default center (Ho Chi Minh City)

    // Generate polylines for the route
    const polylinePositions = locations.map(loc => loc.position);

    const getAnomalyColor = (type) => {
        switch (type) {
            case 'TEMPERATURE':
                return 'volcano';
            case 'HUMIDITY':
                return 'geekblue';
            case 'TIME_GAP':
                return 'gold';
            case 'LOCATION':
                return 'magenta';
            case 'AUTHORIZATION':
                return 'red';
            default:
                return 'purple';
        }
    };

    const getItemIcon = (item) => {
        switch (item.itemType) {
            case 'event':
                return <ClockCircleOutlined style={{ fontSize: '16px' }} />;
            case 'transfer':
                return <SwapOutlined style={{ fontSize: '16px' }} />;
            case 'environment':
                return <ThermometerOutlined style={{ fontSize: '16px' }} />;
            case 'anomaly':
                return <ExclamationCircleOutlined style={{ fontSize: '16px', color: '#f5222d' }} />;
            default:
                return <ClockCircleOutlined style={{ fontSize: '16px' }} />;
        }
    };

    const getTimelineItemColor = (item) => {
        switch (item.itemType) {
            case 'event':
                return 'green';
            case 'transfer':
                return 'blue';
            case 'environment':
                return 'cyan';
            case 'anomaly':
                return 'red';
            default:
                return 'gray';
        }
    };

    const renderTimelineItem = (item) => {
        const formattedDate = moment(item.timestamp).format('MMM DD, YYYY HH:mm:ss');

        let content;

        switch (item.itemType) {
            case 'event':
                content = (
                    <Card size="small" className="timeline-card">
                        <p><strong>{item.type}</strong></p>
                        <p>{item.description}</p>
                        {item.actor && <p><small>Actor: {item.actor}</small></p>}
                        {item.location && (
                            <p>
                                <EnvironmentOutlined />
                                {item.location.address || `${item.location.latitude.toFixed(4)}, ${item.location.longitude.toFixed(4)}`}
                            </p>
                        )}
                        {item.metadata && Object.keys(item.metadata).length > 0 && (
                            <>
                                <Divider style={{ margin: '8px 0' }} />
                                <div className="metadata">
                                    {Object.entries(item.metadata).map(([key, value]) => (
                                        <Tag key={key} color="default">{key}: {typeof value === 'object' ? JSON.stringify(value) : value}</Tag>
                                    ))}
                                </div>
                            </>
                        )}
                    </Card>
                );
                break;

            case 'transfer':
                content = (
                    <Card size="small" className="timeline-card">
                        <p><strong>Transfer</strong></p>
                        <p>From: {item.from}</p>
                        <p>To: {item.to}</p>
                        <p>Status: <Badge status={item.status === 'COMPLETED' ? 'success' : 'processing'} text={item.status} /></p>
                        {item.location && (
                            <p>
                                <EnvironmentOutlined />
                                {item.location.address || `${item.location.latitude.toFixed(4)}, ${item.location.longitude.toFixed(4)}`}
                            </p>
                        )}
                        {item.metadata && Object.keys(item.metadata).length > 0 && (
                            <>
                                <Divider style={{ margin: '8px 0' }} />
                                <div className="metadata">
                                    {Object.entries(item.metadata).map(([key, value]) => (
                                        <Tag key={key} color="default">{key}: {typeof value === 'object' ? JSON.stringify(value) : value}</Tag>
                                    ))}
                                </div>
                            </>
                        )}
                    </Card>
                );
                break;

            case 'environment':
                content = (
                    <Card size="small" className="timeline-card">
                        <p><strong>Environmental Data</strong></p>
                        {item.temperature !== undefined && <p>Temperature: {item.temperature}°C</p>}
                        {item.humidity !== undefined && <p>Humidity: {item.humidity}%</p>}
                        {item.light !== undefined && <p>Light: {item.light} lux</p>}
                        {item.pressure !== undefined && <p>Pressure: {item.pressure} hPa</p>}
                        {item.deviceId && <p>Device: {item.deviceId}</p>}
                        {item.location && (
                            <p>
                                <EnvironmentOutlined />
                                {item.location.address || `${item.location.latitude.toFixed(4)}, ${item.location.longitude.toFixed(4)}`}
                            </p>
                        )}
                    </Card>
                );
                break;

            case 'anomaly':
                content = (
                    <Card size="small" className="timeline-card anomaly-card">
                        <div style={{ display: 'flex', alignItems: 'center', marginBottom: '8px' }}>
                            <WarningOutlined style={{ color: '#f5222d', marginRight: '8px' }} />
                            <strong>Anomaly Detected: {item.type}</strong>
                        </div>
                        <p>{item.description}</p>
                        <p>Confidence: {(item.confidence * 100).toFixed(0)}%</p>
                        {item.expectedValue && item.actualValue && (
                            <>
                                <p>Expected: {item.expectedValue}</p>
                                <p>Actual: {item.actualValue}</p>
                            </>
                        )}
                        {item.relatedEvents && item.relatedEvents.length > 0 && (
                            <p>Related Events: {item.relatedEvents.join(', ')}</p>
                        )}
                        <Divider style={{ margin: '8px 0' }} />
                        <div>
                            <Tag color={getAnomalyColor(item.type)}>{item.type}</Tag>
                        </div>
                    </Card>
                );
                break;

            default:
                content = <p>{JSON.stringify(item)}</p>;
        }

        return (
            <Timeline.Item
                key={item.id}
                color={getTimelineItemColor(item)}
                dot={getItemIcon(item)}
            >
                <div className="timeline-date">{formattedDate}</div>
                {content}
            </Timeline.Item>
        );
    };

    const tabContents = {
        timeline: (
            <div className="timeline-container">
                <Title level={4}>Transaction Timeline</Title>
                <Timeline mode="left">
                    {allTimelineItems.map(renderTimelineItem)}
                </Timeline>
            </div>
        ),
        chart: (
            <div className="chart-container">
                <Title level={4}>Environmental Data</Title>
                {environmentChartData.length > 0 ? (
                    <Line {...config} />
                ) : (
                    <Alert message="No environmental data available" type="info" />
                )}
            </div>
        ),
        map: (
            <div className="map-container">
                <Title level={4}>Geographical Route</Title>
                {locations.length > 0 ? (
                    <div style={{ height: '500px', width: '100%' }}>
                        <MapContainer center={mapCenter} zoom={13} style={{ height: '100%', width: '100%' }}>
                            <TileLayer
                                url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                                attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                            />
                            {locations.map((loc) => (
                                <Marker key={loc.id} position={loc.position}>
                                    <Popup>
                                        <div>
                                            <strong>{loc.type.charAt(0).toUpperCase() + loc.type.slice(1)}</strong>
                                            <p>{loc.description}</p>
                                            <p>{moment(loc.timestamp).format('MMM DD, YYYY HH:mm:ss')}</p>
                                        </div>
                                    </Popup>
                                </Marker>
                            ))}
                            <Polyline positions={polylinePositions} color="blue" weight={3} opacity={0.7} />
                        </MapContainer>
                    </div>
                ) : (
                    <Alert message="No location data available" type="info" />
                )}
            </div>
        ),
        anomalies: (
            <div className="anomalies-container">
                <Title level={4}>Detected Anomalies</Title>
                {timelineData.anomalies.length > 0 ? (
                    <div>
                        {timelineData.anomalies.map((anomaly) => (
                            <Card
                                key={anomaly.id}
                                style={{ marginBottom: '16px' }}
                                title={
                                    <div style={{ display: 'flex', alignItems: 'center' }}>
                                        <ExclamationCircleOutlined style={{ color: '#f5222d', marginRight: '8px' }} />
                                        <span>Anomaly: {anomaly.type}</span>
                                        <Tag color={getAnomalyColor(anomaly.type)} style={{ marginLeft: '8px' }}>
                                            {(anomaly.confidence * 100).toFixed(0)}% confidence
                                        </Tag>
                                    </div>
                                }
                            >
                                <p>{anomaly.description}</p>
                                <Row gutter={16}>
                                    <Col span={12}>
                                        <p><strong>Time:</strong> {moment(anomaly.timestamp).format('MMM DD, YYYY HH:mm:ss')}</p>
                                    </Col>
                                    <Col span={12}>
                                        <p><strong>Type:</strong> {anomaly.type}</p>
                                    </Col>
                                </Row>
                                {anomaly.expectedValue && anomaly.actualValue && (
                                    <Row gutter={16}>
                                        <Col span={12}>
                                            <p><strong>Expected:</strong> {anomaly.expectedValue}</p>
                                        </Col>
                                        <Col span={12}>
                                            <p><strong>Actual:</strong> {anomaly.actualValue}</p>
                                        </Col>
                                    </Row>
                                )}
                                {anomaly.metadata && Object.keys(anomaly.metadata).length > 0 && (
                                    <>
                                        <Divider style={{ margin: '12px 0' }} />
                                        <div className="metadata">
                                            {Object.entries(anomaly.metadata).map(([key, value]) => (
                                                <Tag key={key} color="default">{key}: {typeof value === 'object' ? JSON.stringify(value) : value}</Tag>
                                            ))}
                                        </div>
                                    </>
                                )}
                            </Card>
                        ))}
                    </div>
                ) : (
                    <Alert message="No anomalies detected" type="success" />
                )}
            </div>
        )
    };

    return (
        <div className="transaction-timeline-container">
            <Card
                title={<Title level={3}>Batch Transaction History</Title>}
                extra={
                    <div className="tab-buttons">
                        <Tag
                            color={activeTab === 'timeline' ? 'blue' : 'default'}
                            onClick={() => setActiveTab('timeline')}
                            style={{ cursor: 'pointer' }}
                        >
                            <ClockCircleOutlined /> Timeline
                        </Tag>
                        <Tag
                            color={activeTab === 'chart' ? 'blue' : 'default'}
                            onClick={() => setActiveTab('chart')}
                            style={{ cursor: 'pointer' }}
                        >
                            <ThermometerOutlined /> Environment
                        </Tag>
                        <Tag
                            color={activeTab === 'map' ? 'blue' : 'default'}
                            onClick={() => setActiveTab('map')}
                            style={{ cursor: 'pointer' }}
                        >
                            <AimOutlined /> Map
                        </Tag>
                        <Tag
                            color={activeTab === 'anomalies' ? 'blue' : 'default'}
                            onClick={() => setActiveTab('anomalies')}
                            style={{ cursor: 'pointer' }}
                        >
                            <ExclamationCircleOutlined /> Anomalies
                            {timelineData.anomalies.length > 0 && (
                                <Badge count={timelineData.anomalies.length} style={{ marginLeft: '5px' }} />
                            )}
                        </Tag>
                    </div>
                }
            >
                <div className="batch-info" style={{ marginBottom: '20px' }}>
                    <Row gutter={16}>
                        <Col span={8}>
                            <Text strong>Batch ID:</Text> {timelineData.batchId}
                        </Col>
                        <Col span={8}>
                            <Text strong>Total Events:</Text> {allTimelineItems.length}
                        </Col>
                        <Col span={8}>
                            <Text strong>Anomalies:</Text> {timelineData.anomalies.length}
                        </Col>
                    </Row>
                </div>

                <div className="tab-content">
                    {tabContents[activeTab]}
                </div>
            </Card>

            <style jsx>{`
        .transaction-timeline-container {
          margin: 20px 0;
        }
        .timeline-date {
          font-weight: bold;
          margin-bottom: 8px;
        }
        .timeline-card {
          max-width: 400px;
        }
        .anomaly-card {
          border-left: 2px solid #f5222d;
          background-color: #fff1f0;
        }
        .tab-buttons {
          display: flex;
          gap: 8px;
        }
        .metadata {
          display: flex;
          flex-wrap: wrap;
          gap: 5px;
        }
      `}</style>
        </div>
    );
};

export default TransactionTimeline;
