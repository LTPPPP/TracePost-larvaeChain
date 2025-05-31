'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import {
  FolderPlus,
  LayoutDashboard,
  ShoppingBasket,
  UserRound,
  Mail,
  Phone,
  Calendar,
  Building,
  MapPin
} from 'lucide-react';

import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { useProfile } from '@/hooks/useProfile';

import classNames from 'classnames/bind';
import styles from './Profile.module.scss';

const cx = classNames.bind(styles);

// Types
interface UserInfo {
  avatar: string;
  username: string;
  fullname: string;
  email: string;
  phone: string;
  dob: string;
}

interface CompanyInfo {
  name: string;
  location: string;
  contactInfo: string;
}

function Profile() {
  const { profile, loading, error, refetch } = useProfile();

  const [userInfo, setUserInfo] = useState<UserInfo>({
    avatar: '/img/default-avatar.png',
    username: '',
    fullname: '',
    email: '',
    phone: '',
    dob: ''
  });

  const [companyInfo, setCompanyInfo] = useState<CompanyInfo>({
    name: '',
    location: '',
    contactInfo: ''
  });

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

  // Update user info when profile data changes
  useEffect(() => {
    if (profile) {
      setUserInfo({
        avatar: profile.avatar_url || '/img/default-avatar.png',
        username: profile.username || '',
        fullname: profile.full_name || '',
        email: profile.email || '',
        phone: profile.phone || '',
        dob: profile.date_of_birth ? new Date(profile.date_of_birth).toLocaleDateString() : ''
      });

      // Update company info if available
      if (profile.company) {
        setCompanyInfo({
          name: profile.company.name || '',
          location: profile.company.address || profile.company.location || '',
          contactInfo: profile.company.contact_email || profile.company.phone || ''
        });
      }
    }
  }, [profile]);

  // Loading state
  if (loading) {
    return (
      <div className={cx('wrapper')}>
        <Sidebar menuItems={menuItems} />
        <div className={cx('content')}>
          <div className='flex items-center justify-center h-64'>
            <div className='animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600'></div>
            <span className='ml-2'>Loading profile...</span>
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={cx('wrapper')}>
        <Sidebar menuItems={menuItems} />
        <div className={cx('content')}>
          <div className='flex flex-col items-center justify-center h-64'>
            <p className='text-red-600 mb-4'>Error loading profile: {error}</p>
            <button onClick={refetch} className='bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600'>
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  // No profile data
  if (!profile) {
    return (
      <div className={cx('wrapper')}>
        <Sidebar menuItems={menuItems} />
        <div className={cx('content')}>
          <div className='flex items-center justify-center h-64'>
            <p>No profile data available</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cx('wrapper')}>
      <Sidebar menuItems={menuItems} />

      <div className={cx('content')}>
        <div className='flex justify-between items-center mb-6'>
          <h1 className={cx('page-title')}>Profile</h1>
          <button onClick={refetch} className='bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 text-sm'>
            Refresh
          </button>
        </div>

        <div className={cx('acc-info')}>
          <div className={cx('section-header')}>
            <h2>Personal Info</h2>
            <button className={cx('edit-button')}>EDIT</button>
          </div>

          <div className={cx('user-details')}>
            <div className={cx('avatar-container')}>
              <Image
                src={userInfo.avatar}
                alt='avatar'
                width={120}
                height={120}
                className={cx('avatar')}
                onError={(e) => {
                  const target = e.target as HTMLImageElement;
                  target.src = '/img/default-avatar.png';
                }}
              />
              <button className={cx('change-avatar')}>CHANGE</button>
            </div>

            <div className={cx('info-container')}>
              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Username</span>
                    <span className={cx('value')}>{userInfo.username || 'N/A'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Name</span>
                    <span className={cx('value')}>{userInfo.fullname || 'N/A'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Mail className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Email</span>
                    <span className={cx('value')}>{userInfo.email || 'N/A'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <Phone className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Phone</span>
                    <span className={cx('value')}>{userInfo.phone || 'N/A'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Calendar className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>DOB</span>
                    <span className={cx('value')}>{userInfo.dob || 'N/A'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Role</span>
                    <span className={cx('value')}>{profile.role || 'N/A'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Calendar className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Last Login</span>
                    <span className={cx('value')}>
                      {profile.last_login ? new Date(profile.last_login).toLocaleString() : 'N/A'}
                    </span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Status</span>
                    <span className={cx('value', profile.is_active ? 'text-green-600' : 'text-red-600')}>
                      {profile.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {profile.company && (
          <div className={cx('company-info')}>
            <div className={cx('section-header')}>
              <h2>Company Info</h2>
              <button className={cx('edit-button')}>EDIT</button>
            </div>

            <div className={cx('company-details')}>
              <div className={cx('info-item')}>
                <Building className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Name</span>
                  <span className={cx('value')}>{companyInfo.name || 'N/A'}</span>
                </div>
              </div>

              <div className={cx('info-item')}>
                <MapPin className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Address</span>
                  <span className={cx('value')}>{companyInfo.location || 'N/A'}</span>
                </div>
              </div>

              <div className={cx('info-item')}>
                <Mail className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Contact</span>
                  <span className={cx('value')}>{companyInfo.contactInfo || 'N/A'}</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default Profile;
