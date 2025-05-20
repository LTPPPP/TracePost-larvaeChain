'use client';

import { useEffect, useState } from 'react';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, ShoppingBasket, FolderPlus } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './Workspace.module.scss';
import Clock from '@/components/ui/Clock/Clock';
import HatcheryCard from '@/components/ui/HatcheryCard/HatcheryCard';

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

function Workspace() {
  const [hatchery, setHatchery] = useState<HatcheryData[]>([]);

  // MENU
  const menuItems: MenuItem[] = [
    {
      icon: LayoutDashboard,
      name: 'Workspace',
      link: '/workspace'
    },
    {
      icon: FolderPlus,
      name: 'Create Batch',
      link: '/create-batch'
    },
    {
      icon: ShoppingBasket,
      name: 'Order History',
      link: '/order-history'
    },
    {
      icon: UserRound,
      name: 'Profile',
      link: '/profile'
    }
  ];

  const mockData: HatcheryData[] = [
    {
      id: '1',
      name: 'Pond 1',
      temperature: 28.5,
      ph: 7.8,
      salinity: 15,
      density: 300,
      age: 25
    },
    {
      id: '2',
      name: 'Pond 2',
      temperature: 31.2,
      ph: 7.3,
      salinity: 16,
      density: 280,
      age: 20
    },
    {
      id: '3',
      name: 'Pond 10',
      temperature: 27.8,
      ph: 8.6,
      salinity: 14,
      density: 320,
      age: 30
    }
  ];

  // GET Hatchery
  useEffect(() => {
    const fetchHatcheryData = async () => {
      try {
        // const response = await fetch('/api/hatchery-profile');
        // const data = await response.json();
        // setHatchery(data.hatcheryInfo);

        setHatchery(mockData);
      } catch (error) {
        console.error('Error fetching Hatchery data:', error);
      }
    };

    fetchHatcheryData();
  }, []);

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />
      <div className={cx('container')}>
        <HatcheryCard data={hatchery} />
      </div>
    </div>
  );
}

export default Workspace;
