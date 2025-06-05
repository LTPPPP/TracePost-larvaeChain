'use client';

import { useEffect, useState, useMemo } from 'react';
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
  MapPin,
  RefreshCw
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
  contact_info: string;
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
    contact_info: ''
  });

  // MENU - Dynamic based on user role
  const menuItems = useMemo((): MenuItem[] => {
    const userRole = profile?.role?.toLowerCase();

    switch (userRole) {
      case 'hatchery':
      case 'admin':
        return [
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

      case 'user':
        return [
          {
            icon: LayoutDashboard,
            name: 'Company List',
            link: '/company-list'
          },
          {
            icon: UserRound,
            name: 'Profile',
            link: '/profile'
          }
        ];

      case 'distributor':
        return [
          {
            icon: LayoutDashboard,
            name: 'Dashboard',
            link: '/distributor'
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

      default:
        return [
          {
            icon: LayoutDashboard,
            name: 'Dashboard',
            link: '/dashboard'
          },
          {
            icon: UserRound,
            name: 'Profile',
            link: '/profile'
          }
        ];
    }
  }, [profile?.role]);

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

      if (profile.company) {
        setCompanyInfo({
          name: profile.company.name || '',
          location: profile.company.location || '',
          contact_info: profile.company.contact_info || ''
        });
      }
    }
  }, [profile]);

  if (loading) {
    return (
      <div className={cx('wrapper')}>
        <Sidebar menuItems={menuItems} />
        <div className={cx('content')}>
          <div className={cx('loading-container')}>
            <div className={cx('loading-spinner')}></div>
            <span className={cx('loading-text')}>Loading profile...</span>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cx('wrapper')}>
        <Sidebar menuItems={menuItems} />
        <div className={cx('content')}>
          <div className={cx('error-container')}>
            <p className={cx('error-text')}>
              Oops! Something went wrong while loading your profile.
              <br />
              <strong>Error:</strong> {error}
            </p>
            <button onClick={refetch} className={cx('retry-button')}>
              <RefreshCw size={16} style={{ marginRight: '8px', display: 'inline-block' }} />
              Try Again
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
          <div className={cx('error-container')}>
            <p className={cx('error-text')}>No profile data available</p>
            <button onClick={refetch} className={cx('retry-button')}>
              <RefreshCw size={16} style={{ marginRight: '8px', display: 'inline-block' }} />
              Refresh
            </button>
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
          <h1 className={cx('page-title')}>My Profile</h1>
          <button onClick={refetch} className={cx('refresh-button')}>
            <RefreshCw size={16} style={{ marginRight: '6px', display: 'inline-block' }} />
            Refresh
          </button>
        </div>

        <div className={cx('acc-info')}>
          <div className={cx('section-header')}>
            <h2>Personal Information</h2>
            <button className={cx('edit-button')}>EDIT PROFILE</button>
          </div>

          <div className={cx('user-details')}>
            <div className={cx('avatar-container')}>
              <Image
                src={userInfo.avatar}
                alt='Profile Avatar'
                width={140}
                height={140}
                className={cx('avatar')}
                onError={(e) => {
                  const target = e.target as HTMLImageElement;
                  target.src = '/img/default-avatar.png';
                }}
              />
              <button className={cx('change-avatar')}>CHANGE PHOTO</button>
            </div>

            <div className={cx('info-container')}>
              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Username</span>
                    <span className={cx('value')}>{userInfo.username || 'Not provided'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Full Name</span>
                    <span className={cx('value')}>{userInfo.fullname || 'Not provided'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Mail className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Email Address</span>
                    <span className={cx('value')}>{userInfo.email || 'Not provided'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <Phone className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Phone Number</span>
                    <span className={cx('value')}>{userInfo.phone || 'Not provided'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Calendar className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Date of Birth</span>
                    <span className={cx('value')}>{userInfo.dob || 'Not provided'}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Role</span>
                    <span className={cx('value')}>{profile.role || 'Not assigned'}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Calendar className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Last Login</span>
                    <span className={cx('value')}>
                      {profile.last_login ? new Date(profile.last_login).toLocaleString() : 'Never logged in'}
                    </span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Account Status</span>
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
              <h2>Company Information</h2>
              <button className={cx('edit-button')}>EDIT COMPANY</button>
            </div>

            <div className={cx('company-details')}>
              <div className={cx('info-item')}>
                <Building className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Company Name</span>
                  <span className={cx('value')}>{companyInfo.name || 'Not provided'}</span>
                </div>
              </div>

              <div className={cx('info-item')}>
                <MapPin className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Address</span>
                  <span className={cx('value')}>{companyInfo.location || 'Not provided'}</span>
                </div>
              </div>

              <div className={cx('info-item')}>
                <Mail className={cx('icon')} />
                <div className={cx('info-content')}>
                  <span className={cx('label')}>Contact Information</span>
                  <span className={cx('value')}>{companyInfo.contact_info || 'Not provided'}</span>
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
