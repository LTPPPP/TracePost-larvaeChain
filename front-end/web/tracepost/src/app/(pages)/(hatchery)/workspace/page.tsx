'use client';

import { useEffect, useState } from 'react';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, FolderPlus } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './Workspace.module.scss';
import Clock from '@/components/ui/Clock/Clock';
import HatcheryCard from '@/components/ui/HatcheryCard/HatcheryCard';
import { getProfile } from '@/api/profile';
import { getCompanyBatchesWithEnvironment } from '@/api/company';

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

interface UserProfile {
  id: number;
  username: string;
  email: string;
  role: string;
  company_id: number;
  company?: {
    id: number;
    name: string;
    location: string;
    contact_info: string;
  };
}

function Workspace() {
  const [hatchery, setHatchery] = useState<HatcheryData[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);

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
      icon: UserRound,
      name: 'Profile',
      link: '/profile'
    }
  ];

  // GET Profile and Company Batches
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        const profileResponse = await getProfile();

        if (!profileResponse.success || !profileResponse.data) {
          throw new Error('Failed to fetch user profile');
        }

        const profile = profileResponse.data as UserProfile;
        setUserProfile(profile);

        localStorage.setItem('userInfo', JSON.stringify(profile));

        if (!profile.company_id) {
          setError('User is not associated with any company');
          return;
        }

        const batchesResult = await getCompanyBatchesWithEnvironment(profile.company_id);

        if (!batchesResult.success) {
          throw new Error(batchesResult.error || 'Failed to fetch company batches');
        }

        setHatchery(batchesResult.data);
      } catch (error) {
        console.error('Error fetching workspace data:', error);
        setError(error instanceof Error ? error.message : 'Error fetching data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // Handle environment update
  const handleUpdateEnvironment = (id: string, updatedData: Partial<HatcheryData>) => {
    setHatchery((prevData) => prevData.map((item) => (item.id === id ? { ...item, ...updatedData } : item)));
  };

  if (loading) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('container')}>
          <div className={cx('loading')}>Loading workspace data...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('container')}>
          <div className={cx('error')}>
            <h3>Error</h3>
            <p>{error}</p>
            <button onClick={() => window.location.reload()} className={cx('retry-button')}>
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />

      <div className={cx('container')}>
        {userProfile && (
          <div className={cx('workspace-header')}>
            <h1 className={cx('workspace-title')}>Workspace - {userProfile.company?.name || 'Your Company'}</h1>
            <p className={cx('workspace-subtitle')}>
              Welcome back, {userProfile.username}! Here are company active batches.
              {userProfile.role === 'hatchery' && (
                <span className={cx('edit-hint')}> Click on any card to edit environment data.</span>
              )}
            </p>
          </div>
        )}

        {hatchery.length > 0 ? (
          <HatcheryCard data={hatchery} onUpdateEnvironment={handleUpdateEnvironment} />
        ) : (
          <div className={cx('no-data')}>
            <h3>No Active Batches</h3>
            <p>There are currently no active batches for your company.</p>
            <a href='/create-batch' className={cx('create-batch-link')}>
              Create New Batch
            </a>
          </div>
        )}
      </div>
    </div>
  );
}

export default Workspace;
