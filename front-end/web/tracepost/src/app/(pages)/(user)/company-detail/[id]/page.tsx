'use client';

import { useEffect, useState } from 'react';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, ShoppingBasket, MapPin, Mail, LandPlot } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './CompanyDetail.module.scss';
import Clock from '@/components/ui/Clock/Clock';
import HatcheryCard from '@/components/ui/HatcheryCard/HatcheryCard';
import { useParams } from 'next/navigation';
import Image from 'next/image';

const cx = classNames.bind(styles);

interface CompanyData {
  id: string;
  name: string;
  address: string;
  contact: string;
  totalPond: number;
  certificate: string;
}

interface HatcheryData {
  id: string;
  name: string;
  temperature: number;
  ph: number;
  salinity: number;
  density: number;
  age: number;
}

function CompanyDetail() {
  const { id } = useParams();
  const [company, setCompany] = useState<CompanyData>({
    id: '1',
    name: 'ABC Corporation',
    address: 'London City, England',
    contact: 'contact@abccorp.com',
    totalPond: 3,
    certificate: '/img/default-certificate.png'
  });
  const [hatchery, setHatchery] = useState<HatcheryData[]>([]);

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
        // const response = await fetch('/api/hatchery-profile/{companyId}');
        // const data = await response.json();
        // setHatchery(data.hatcheryInfo);

        setHatchery(mockData);
      } catch (error) {
        console.error('Error fetching Hatchery data:', error);
      }
    };

    fetchHatcheryData();
  }, []);

  // GET Company By ID
  useEffect(() => {
    const fetchCompanyData = async () => {
      try {
        // const response = await fetch(`/api/company-list/${id}`);
        // const data = await response.json();
        // setCompany(data.companyInfo);
      } catch (error) {
        console.error('Error fetching Company data:', error);
      }
    };

    // fetchCompanyData();
  }, [id]);

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />

      <div className={cx('company-info')}>
        <div className={cx('company-content')}>
          <div className={cx('company-header')}>
            <h1 className={cx('company-name')}>{company.name}</h1>
          </div>

          <div className={cx('company-details')}>
            <div className={cx('detail')}>
              <div className={cx('icon')}>
                <MapPin size={25} />
              </div>
              <div className={cx('detail-content')}>
                <div className={cx('detail-label')}>Address</div>
                <div className={cx('detail-value')}>{company.address}</div>
              </div>
            </div>

            <div className={cx('detail')}>
              <div className={cx('icon')}>
                <Mail size={25} />
              </div>
              <div className={cx('detail-content')}>
                <div className={cx('detail-label')}>Contact</div>
                <div className={cx('detail-value')}>{company.contact}</div>
              </div>
            </div>

            <div className={cx('detail')}>
              <div className={cx('icon')}>
                <LandPlot size={25} />
              </div>
              <div className={cx('detail-content')}>
                <div className={cx('detail-label')}>Total Pond</div>
                <div className={cx('detail-value')}>{company.totalPond}</div>
              </div>
            </div>
          </div>
        </div>

        <div className={cx('certificate')}>
          <div className={cx('certificate-img')}>
            <Image src={company.certificate} alt='certificate' width={400} height={300} />
          </div>
          <span className={cx('certificate-text')}>Operating Certificate</span>
        </div>
      </div>

      <div className={cx('container')}>
        <HatcheryCard data={hatchery} />
      </div>
    </div>
  );
}

export default CompanyDetail;
