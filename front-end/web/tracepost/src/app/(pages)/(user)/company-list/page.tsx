'use client';

import { useEffect, useState } from 'react';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, ShoppingBasket, MapPin, LandPlot, Mail } from 'lucide-react';
import { getListCompany, ApiCompany } from '@/api/company';
import { getListHatcheries, countHatcheriesByCompany } from '@/api/hatchery';

import classNames from 'classnames/bind';
import styles from './CompanyList.module.scss';
import Clock from '@/components/ui/Clock/Clock';
import Link from 'next/link';

const cx = classNames.bind(styles);

interface CompanyData {
  id: number;
  name: string;
  location: string;
  contact_info: string;
  totalPond: number;
}

function CompanyList() {
  const [companies, setCompanies] = useState<CompanyData[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // MENU
  const menuItems: MenuItem[] = [
    {
      icon: LayoutDashboard,
      name: 'Company List',
      link: '/company-list'
    },
    // {
    //   icon: ShoppingBasket,
    //   name: 'Order History',
    //   link: '/order-history'
    // },
    {
      icon: UserRound,
      name: 'Profile',
      link: '/profile'
    }
  ];

  // GET Company Data
  useEffect(() => {
    const fetchCompanyData = async () => {
      try {
        setLoading(true);
        setError(null);

        const [companiesResponse, hatcheriesResponse] = await Promise.all([getListCompany(), getListHatcheries()]);

        if (companiesResponse.success && hatcheriesResponse.success) {
          const processedCompanies: CompanyData[] = companiesResponse.data.map((company: ApiCompany) => ({
            id: company.id,
            name: company.name,
            location: company.location,
            contact_info: company.contact_info,
            totalPond: countHatcheriesByCompany(
              Array.isArray(hatcheriesResponse.data) ? hatcheriesResponse.data.flat() : hatcheriesResponse.data,
              company.id
            )
          }));

          setCompanies(processedCompanies);
        } else {
          setError('Failed to fetch company data');
        }
      } catch (error) {
        console.error('Error fetching Company data:', error);
        setError('Error fetching company data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchCompanyData();
  }, []);

  if (loading) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('company-list')}>
          <div className={cx('loading')}>Loading companies...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('company-list')}>
          <div className={cx('error')}>{error}</div>
        </div>
      </div>
    );
  }

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />

      <div className={cx('company-list')}>
        {companies.length === 0 ? (
          <div className={cx('no-data')}>No companies found</div>
        ) : (
          companies.map((company) => (
            <Link href={'/company-detail/' + company.id} key={company.id} className={cx('company-item')}>
              <div className={cx('company-name')}>{company.name}</div>
              <div className={cx('company-details')}>
                <div className={cx('detail')}>
                  <MapPin size={25} />
                  <div className={cx('detail-content')}>
                    <div className={cx('detail-label')}>Address</div>
                    <div className={cx('detail-value')}>{company.location}</div>
                  </div>
                </div>

                <div className={cx('detail')}>
                  <Mail size={25} />
                  <div className={cx('detail-content')}>
                    <div className={cx('detail-label')}>Contact</div>
                    <div className={cx('detail-value')}>{company.contact_info}</div>
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
          ))
        )}
      </div>
    </div>
  );
}

export default CompanyList;
