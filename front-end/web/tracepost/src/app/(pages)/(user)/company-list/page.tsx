'use client';

import { useEffect, useState } from 'react';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, ShoppingBasket, MapPin, LandPlot, Mail } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './CompanyList.module.scss';
import Clock from '@/components/ui/Clock/Clock';
import Link from 'next/link';

const cx = classNames.bind(styles);

interface CompanyData {
  id: string;
  name: string;
  address: string;
  contact: string;
  totalPond: number;
}

function CompanyList() {
  const [companies, setCompanies] = useState<CompanyData[]>([
    {
      id: '1',
      name: 'ABC Corporation',
      address: 'London City, England',
      contact: 'contact@abccorp.com',
      totalPond: 2
    },
    {
      id: '2',
      name: 'APK Corporation',
      address: 'London City, England',
      contact: 'contact@abccorp.com',
      totalPond: 4
    },
    {
      id: '3',
      name: 'KFC Corporation',
      address: 'London City, England',
      contact: 'contact@abccorp.com',
      totalPond: 5
    },
    {
      id: '4',
      name: 'KFC Corporation',
      address: 'London City, England',
      contact: 'contact@abccorp.com',
      totalPond: 5
    },
    {
      id: '5',
      name: 'KFC Corporation',
      address: 'London City, England',
      contact: 'contact@abccorp.com',
      totalPond: 5
    }
  ]);

  // MENU
  const menuItems: MenuItem[] = [
    {
      icon: LayoutDashboard,
      name: 'Company List',
      link: '/company-list'
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

  // GET Company
  useEffect(() => {
    const fetchCompanyData = async () => {
      try {
        // const response = await fetch('/api/company-list');
        // const data = await response.json();
        // setCompanies(data.companyInfo);
      } catch (error) {
        console.error('Error fetching Company data:', error);
      }
    };

    // fetchCompanyData();
  }, []);

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />

      <div className={cx('company-list')}>
        {companies.map((company) => (
          <Link href={'/company-detail/' + company.id} key={company.id} className={cx('company-item')}>
            <div className={cx('company-name')}>{company.name}</div>
            <div className={cx('company-details')}>
              <div className={cx('detail')}>
                <MapPin size={25} />

                <div className={cx('detail-content')}>
                  <div className={cx('detail-label')}>Address</div>
                  <div className={cx('detail-value')}>{company.address}</div>
                </div>
              </div>

              <div className={cx('detail')}>
                <Mail size={25} />

                <div className={cx('detail-content')}>
                  <div className={cx('detail-label')}>Contact</div>
                  <div className={cx('detail-value')}>{company.contact}</div>
                </div>
              </div>

              <div className={cx('detail')}>
                <LandPlot size={25} />
                <div className={cx('detail-content')}>
                  <div className={cx('detail-label')}>Total Pond</div>
                  <div className={cx('detail-value')}>{company.totalPond}</div>
                </div>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}

export default CompanyList;
