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
  const [userInfo, setUserInfo] = useState<UserInfo>({
    avatar: '/img/default-avatar.png',
    username: '@kesatnhan',
    fullname: 'Jack The ripper',
    email: 'Jack.kl@gmail.com',
    phone: '+84 123 456 789',
    dob: '01/01/1990'
  });

  const [companyInfo, setCompanyInfo] = useState<CompanyInfo>({
    name: 'ABC Corporation',
    location: 'London City, England',
    contactInfo: 'contact@abccorp.com'
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

  // GET PROFILE
  useEffect(() => {
    const fetchUserData = async () => {
      try {
        // const response = await fetch('/api/user-profile');
        // const data = await response.json();
        // setUserInfo(data.userInfo);
        // setCompanyInfo(data.companyInfo);
      } catch (error) {
        console.error('Error fetching user data:', error);
      }
    };

    // fetchUserData();
  }, []);

  return (
    <div className={cx('wrapper')}>
      <Sidebar menuItems={menuItems} />

      <div className={cx('content')}>
        <h1 className={cx('page-title')}>Profile</h1>

        <div className={cx('acc-info')}>
          <div className={cx('section-header')}>
            <h2>Personal Info</h2>
            <button className={cx('edit-button')}>EDIT</button>
          </div>

          <div className={cx('user-details')}>
            <div className={cx('avatar-container')}>
              <Image src={userInfo.avatar} alt='avatar' width={120} height={120} className={cx('avatar')} />
              <button className={cx('change-avatar')}>CHANGE</button>
            </div>

            <div className={cx('info-container')}>
              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Username</span>
                    <span className={cx('value')}>{userInfo.username}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <UserRound className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Name</span>
                    <span className={cx('value')}>{userInfo.fullname}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Mail className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Email</span>
                    <span className={cx('value')}>{userInfo.email}</span>
                  </div>
                </div>

                <div className={cx('info-item')}>
                  <Phone className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>Phone</span>
                    <span className={cx('value')}>{userInfo.phone}</span>
                  </div>
                </div>
              </div>

              <div className={cx('info-group')}>
                <div className={cx('info-item')}>
                  <Calendar className={cx('icon')} />
                  <div className={cx('info-content')}>
                    <span className={cx('label')}>DOB</span>
                    <span className={cx('value')}>{userInfo.dob}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

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
                <span className={cx('value')}>{companyInfo.name}</span>
              </div>
            </div>

            <div className={cx('info-item')}>
              <MapPin className={cx('icon')} />
              <div className={cx('info-content')}>
                <span className={cx('label')}>Address</span>
                <span className={cx('value')}>{companyInfo.location}</span>
              </div>
            </div>

            <div className={cx('info-item')}>
              <Mail className={cx('icon')} />
              <div className={cx('info-content')}>
                <span className={cx('label')}>Contact</span>
                <span className={cx('value')}>{companyInfo.contactInfo}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Profile;
