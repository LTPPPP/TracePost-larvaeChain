'use client';

import { useState, useEffect } from 'react';
import classNames from 'classnames/bind';
import styles from './Clock.module.scss';

const cx = classNames.bind(styles);

const Clock: React.FC = () => {
  const [dateTime, setDateTime] = useState<Date>(new Date());

  useEffect(() => {
    const timer = setInterval(() => {
      setDateTime(new Date());
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const getDayName = (): string => {
    const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    return days[dateTime.getDay()];
  };

  const formatTime = (): string => {
    const hours = dateTime.getHours().toString().padStart(2, '0');
    const minutes = dateTime.getMinutes().toString().padStart(2, '0');
    const seconds = dateTime.getSeconds().toString().padStart(2, '0');
    return `${hours}:${minutes}:${seconds}`;
  };

  return (
    <div className={cx('wrapper')}>
      <div className={cx('clock')}>
        {getDayName()} {formatTime()}
      </div>
    </div>
  );
};

export default Clock;
