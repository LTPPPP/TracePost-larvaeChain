'use client';

import { Thermometer, Droplet, Waves, Clock, Shrimp, Sigma } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './HatcheryCard.module.scss';

const cx = classNames.bind(styles);

interface HatcheryData {
  id: string;
  name: string;
  temperature: number;
  ph: number;
  salinity: number;
  density: number;
  age: number;
  species: string;
  quantity: number;
  status?: string;
  batchId?: number;
  hatcheryId?: number;
}

interface HatcheryCardProps {
  data: HatcheryData[];
}

function HatcheryCard({ data }: HatcheryCardProps) {
  // Khoảng ổn định
  const TEMP_MIN = 28;
  const TEMP_MAX = 30;
  const PH_MIN = 7.5;
  const PH_MAX = 8.5;

  /**
   * LOW   - BLUE
   * STABLE- GREEN
   * HIGH  - RED
   */
  const getTemperatureColor = (temp: number) => {
    if (temp < TEMP_MIN) return 'blue';
    if (temp > TEMP_MAX) return 'red';
    return 'green';
  };

  const getPhColor = (ph: number) => {
    if (ph < PH_MIN) return 'blue';
    if (ph > PH_MAX) return 'red';
    return 'green';
  };

  return (
    <div className={cx('environment-container')}>
      <div className={cx('section-header')}>
        <h2 className={cx('section-title')}>Active Batches ({data.length})</h2>
      </div>

      <div className={cx('ponds-grid')}>
        {data.map((pond) => (
          <div key={pond.id} className={cx('pond-card')}>
            <div className={cx('pond-header')}>
              <div className={cx('pond-title')}>
                <h3 className={cx('pond-name')}>
                  {pond.name} - {pond.species}
                </h3>
              </div>
            </div>

            <div className={cx('pond-details')}>
              <div className={cx('detail-item')}>
                <Sigma className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Quantity:</span>
                <span className={cx('detail-value')}>{pond.quantity.toLocaleString()}</span>
              </div>

              <div className={cx('detail-item')}>
                <Thermometer className={cx('detail-icon', getTemperatureColor(pond.temperature))} size={25} />
                <span className={cx('detail-label')}>Temp:</span>
                <span className={cx('detail-value', getTemperatureColor(pond.temperature))}>
                  {pond.temperature > 0 ? `${pond.temperature}°C` : 'N/A'}
                </span>
              </div>

              <div className={cx('detail-item')}>
                <Droplet className={cx('detail-icon', getPhColor(pond.ph))} size={25} />
                <span className={cx('detail-label')}>pH:</span>
                <span className={cx('detail-value', getPhColor(pond.ph))}>{pond.ph > 0 ? pond.ph : 'N/A'}</span>
              </div>

              <div className={cx('detail-item')}>
                <Waves className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Salinity:</span>
                <span className={cx('detail-value')}>{pond.salinity > 0 ? `${pond.salinity} ppt` : 'N/A'}</span>
              </div>

              <div className={cx('detail-item')}>
                <Shrimp className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Density:</span>
                <span className={cx('detail-value')}>{pond.density > 0 ? `${pond.density}/m³` : 'N/A'}</span>
              </div>

              <div className={cx('detail-item')}>
                <Clock className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Age:</span>
                <span className={cx('detail-value')}>{pond.age > 0 ? `${pond.age} days` : 'N/A'}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default HatcheryCard;
