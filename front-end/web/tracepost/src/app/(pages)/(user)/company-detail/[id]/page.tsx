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
import { getCompanyById } from '@/api/company';
import { ApiHatchery } from '@/api/hatchery';
import { ApiBatch, ApiEnvironment, getBatches, getEnvironment } from '@/api/batch';

const cx = classNames.bind(styles);

interface CompanyData {
  id: number;
  name: string;
  address: string;
  contact: string;
  hatcheries: ApiHatchery[];
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
  species: string;
  quantity: number;
}

function CompanyDetail() {
  const { id } = useParams();
  const companyId = Array.isArray(id) ? parseInt(id[0]) : parseInt(id as string);

  const [company, setCompany] = useState<CompanyData | null>(null);
  const [hatchery, setHatchery] = useState<HatcheryData[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

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

  // GET Company and Hatchery
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        //Company data
        const companyResponse = await getCompanyById(companyId);
        if (companyResponse.success && companyResponse.data) {
          const companyData = companyResponse.data;
          setCompany({
            id: companyData.id,
            name: companyData.name,
            address: companyData.location,
            contact: companyData.contact_info,
            hatcheries: companyData.hatcheries || [],
            totalPond: companyData.hatcheries?.length || 0,
            certificate: '/img/default-certificate.png'
          });
        }

        const batchesResponse = await getBatches();
        if (batchesResponse.success && batchesResponse.data) {
          const batchesData = Array.isArray(batchesResponse.data) ? batchesResponse.data : [batchesResponse.data];

          const companyBatches = batchesData.filter((batch: ApiBatch) => batch.hatchery?.company_id === companyId);

          const environmentPromises = companyBatches.map(async (batch: ApiBatch): Promise<HatcheryData | null> => {
            try {
              const envResponse = await getEnvironment(batch.id);
              if (envResponse.success && Array.isArray(envResponse.data) && envResponse.data.length > 0) {
                const envData = envResponse.data[0] as ApiEnvironment;
                return {
                  id: batch.id.toString(),
                  name: envData.facility_info?.hatchery_name || batch.hatchery?.name || 'Unknown Hatchery',
                  temperature: envData.temperature ?? 0,
                  ph: envData.ph ?? 0,
                  salinity: envData.salinity ?? 0,
                  density: envData.density ?? 0,
                  age: envData.age ?? 0,
                  species: batch.species || 'Unknown Species',
                  quantity: batch.quantity || 0
                } as HatcheryData;
              }
              return null;
            } catch (error) {
              console.error(`Error fetching environment for batch ${batch.id}:`, error);
              return null;
            }
          });

          const environmentResults = await Promise.all(environmentPromises);
          const validEnvironments = environmentResults.filter((env): env is HatcheryData => env !== null);

          setHatchery(validEnvironments);
        }
      } catch (error) {
        console.error('Error fetching data:', error);
        setError('Error fetching data. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    if (companyId) {
      fetchData();
    }
  }, [companyId]);

  if (loading) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('company-info')}>
          <div className={cx('loading')}>Loading company data...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('company-info')}>
          <div className={cx('error')}>{error}</div>
        </div>
      </div>
    );
  }

  if (!company) {
    return (
      <div className={cx('wrapper')}>
        <Clock />
        <Sidebar menuItems={menuItems} />
        <div className={cx('company-info')}>
          <div className={cx('error')}>Company not found</div>
        </div>
      </div>
    );
  }

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
