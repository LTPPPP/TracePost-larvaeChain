'use client';

import { Thermometer, Droplet, Waves, Clock, Shrimp } from 'lucide-react';

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
      <h2 className={cx('section-title')}>Environment ({data.length})</h2>

      <div className={cx('ponds-grid')}>
        {data.map((pond) => (
          <div key={pond.id} className={cx('pond-card')}>
            <div className={cx('pond-header')}>
              <h3 className={cx('pond-name')}>{pond.name}</h3>
            </div>

            <div className={cx('pond-details')}>
              <div className={cx('detail-item')}>
                <Thermometer className={cx('detail-icon', getTemperatureColor(pond.temperature))} size={25} />
                <span className={cx('detail-label')}>Temp:</span>
                <span className={cx('detail-value')}>{pond.temperature}°C</span>
              </div>

              <div className={cx('detail-item')}>
                <Droplet className={cx('detail-icon', getPhColor(pond.ph))} size={25} />
                <span className={cx('detail-label')}>pH:</span>
                <span className={cx('detail-value')}>{pond.ph}</span>
              </div>

              <div className={cx('detail-item')}>
                <Waves className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Salinity:</span>
                <span className={cx('detail-value')}>{pond.salinity} ppt</span>
              </div>

              <div className={cx('detail-item')}>
                <Shrimp className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Density:</span>
                <span className={cx('detail-value')}>{pond.density}/m³</span>
              </div>

              <div className={cx('detail-item')}>
                <Clock className={cx('detail-icon')} size={25} />
                <span className={cx('detail-label')}>Age:</span>
                <span className={cx('detail-value')}>{pond.age} days</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default HatcheryCard;
