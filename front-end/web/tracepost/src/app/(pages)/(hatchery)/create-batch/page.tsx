'use client';

import { useState, useEffect } from 'react';
import {
  Building2,
  Factory,
  Fish,
  Sigma,
  Plus,
  ArrowRight,
  LayoutDashboard,
  FolderPlus,
  UserRound
} from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './CreateBatch.module.scss';

import { getListCompany, createHatchery, ApiCompany } from '@/api/company';
import { getListHatcheries, ApiHatchery } from '@/api/hatchery';
import { createBatch } from '@/api/batch';
import { getProfile } from '@/api/profile';
import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import Clock from '@/components/ui/Clock/Clock';

const cx = classNames.bind(styles);

interface FormData {
  // Hatchery data
  hatchery_name: string;

  // Batch data
  species: string;
  quantity: number;
}

function CreateBatch() {
  const [userCompany, setUserCompany] = useState<ApiCompany | null>(null);
  const [hatcheries, setHatcheries] = useState<ApiHatchery[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<FormData>({
    hatchery_name: '',
    species: '',
    quantity: 0
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
      icon: UserRound,
      name: 'Profile',
      link: '/profile'
    }
  ];

  useEffect(() => {
    loadUserCompany();
    loadHatcheries();
  }, []);

  const loadUserCompany = async () => {
    try {
      const profileResponse = await getProfile();
      if (profileResponse.success) {
        const companyId = profileResponse.data.company_id;

        if (companyId) {
          const companiesResponse = await getListCompany();
          if (companiesResponse.success && Array.isArray(companiesResponse.data)) {
            const company = companiesResponse.data.find((c) => c.id === companyId);
            if (company) {
              setUserCompany(company);
            }
          }
        }
      }
    } catch (error) {
      console.error('Error loading user company:', error);
    }
  };

  const loadHatcheries = async () => {
    try {
      const response = await getListHatcheries();
      if (response.success && Array.isArray(response.data)) {
        setHatcheries(response.data);
      }
    } catch (error) {
      console.error('Error loading hatcheries:', error);
    }
  };

  const handleInputChange = (field: keyof FormData, value: string | number) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!userCompany) {
      return;
    }

    setLoading(true);

    try {
      console.log('=== Creating Batch Process Started ===');

      // Step 1: Create Hatchery
      console.log('Step 1: Creating hatchery with data:', {
        company_id: userCompany.id,
        name: formData.hatchery_name
      });

      const hatcheryResponse = await createHatchery({
        company_id: userCompany.id,
        name: formData.hatchery_name
      });

      console.log('Hatchery response:', hatcheryResponse);

      if (!hatcheryResponse.success) {
        throw new Error(`Failed to create hatchery: ${hatcheryResponse.message || 'Unknown error'}`);
      }

      // Kiểm tra cấu trúc response từ hatchery
      const hatcheryId = hatcheryResponse.data?.id || hatcheryResponse.data;
      if (!hatcheryId) {
        console.error('Invalid hatchery response structure:', hatcheryResponse);
        throw new Error('Invalid hatchery response - missing ID');
      }

      console.log('Hatchery created successfully with ID:', hatcheryId);

      // Step 2: Create Batch
      const batchData = {
        hatchery_id: hatcheryId,
        quantity: formData.quantity,
        species: formData.species
      };

      console.log('Step 2: Creating batch with data:', batchData);

      const batchResponse = await createBatch(batchData);
      console.log('Batch response:', batchResponse);

      if (!batchResponse.success) {
        throw new Error(`Failed to create batch: ${batchResponse.message || 'Unknown error'}`);
      }

      console.log('=== Batch Creation Process Completed Successfully ===');

      setFormData({
        hatchery_name: '',
        species: '',
        quantity: 0
      });
      setCurrentStep(1);
    } catch (error) {
      console.error('=== Error creating batch ===', error);

      if (error instanceof Error) {
        console.error('Error message:', error.message);
        console.error('Error stack:', error.stack);
      } else {
        console.error('Unknown error:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  const getStepStatus = (step: number) => {
    if (step < currentStep) return 'completed';
    if (step === currentStep) return 'active';
    return 'pending';
  };

  const nextStep = () => {
    if (currentStep < 2) {
      setCurrentStep(currentStep + 1);
    }
  };

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  const canProceedStep1 = userCompany && formData.hatchery_name.trim();
  const canProceedStep2 = formData.species.trim() && formData.quantity > 0;

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />

      <div className={cx('container')}>
        <div className={cx('header')}>
          <h1 className={cx('title')}>Create New Batch</h1>
          <p className={cx('subtitle')}>Set up your aquaculture batch in 2 simple steps</p>
        </div>

        <div className={cx('progress-steps')}>
          <div className={cx('step', getStepStatus(1))}>
            <div className={cx('step-icon')}>
              <Factory size={20} />
            </div>
            <span className={cx('step-label')}>Hatchery</span>
          </div>
          <ArrowRight className={cx('step-arrow')} size={16} />
          <div className={cx('step', getStepStatus(2))}>
            <div className={cx('step-icon')}>
              <Fish size={20} />
            </div>
            <span className={cx('step-label')}>Batch Info</span>
          </div>
        </div>

        <form onSubmit={handleSubmit} className={cx('form')}>
          {/* Step 1: Hatchery */}
          {currentStep === 1 && (
            <div className={cx('step-content')}>
              <div className={cx('step-header')}>
                <Factory className={cx('step-icon-large')} size={32} />
                <h2>Hatchery Setup</h2>
                <p>Create hatchery for your company</p>
              </div>

              <div className={cx('form-grid')}>
                <div className={cx('form-group')}>
                  <label className={cx('label')}>
                    <Building2 size={18} />
                    Your Company
                  </label>
                  <div className={cx('company-info')}>
                    {userCompany ? (
                      <div className={cx('company-display')}>
                        <span className={cx('company-name')}>{userCompany.name}</span>
                        <span className={cx('company-type')}>({userCompany.type})</span>
                        <span className={cx('company-location')}>{userCompany.location}</span>
                      </div>
                    ) : (
                      <div className={cx('company-loading')}>Loading company information...</div>
                    )}
                  </div>
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('label')}>
                    <Factory size={18} />
                    Hatchery Name
                  </label>
                  <input
                    type='text'
                    value={formData.hatchery_name}
                    onChange={(e) => handleInputChange('hatchery_name', e.target.value)}
                    className={cx('input')}
                    placeholder='Enter hatchery name'
                    required
                  />
                </div>
              </div>

              <div className={cx('step-actions')}>
                <button
                  type='button'
                  onClick={nextStep}
                  disabled={!canProceedStep1}
                  className={cx('btn', 'btn-primary')}
                >
                  Next Step <ArrowRight size={16} />
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Batch Info */}
          {currentStep === 2 && (
            <div className={cx('step-content')}>
              <div className={cx('step-header')}>
                <Fish className={cx('step-icon-large')} size={32} />
                <h2>Batch Information</h2>
                <p>Define species and quantity</p>
              </div>

              <div className={cx('form-grid')}>
                <div className={cx('form-group')}>
                  <label className={cx('label')}>
                    <Fish size={18} />
                    Species
                  </label>
                  <input
                    type='text'
                    value={formData.species}
                    onChange={(e) => handleInputChange('species', e.target.value)}
                    className={cx('input')}
                    placeholder='e.g., Whiteleg Shrimp, Tiger Prawn'
                    required
                  />
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('label')}>
                    <Sigma size={18} />
                    Quantity
                  </label>
                  <input
                    type='number'
                    value={formData.quantity}
                    onChange={(e) => handleInputChange('quantity', parseInt(e.target.value))}
                    className={cx('input')}
                    placeholder='Number of individuals'
                    min='1'
                    required
                  />
                </div>
              </div>

              <div className={cx('step-actions')}>
                <button type='button' onClick={prevStep} className={cx('btn', 'btn-secondary')}>
                  Previous
                </button>
                <button type='submit' disabled={!canProceedStep2 || loading} className={cx('btn', 'btn-success')}>
                  {loading ? (
                    'Creating...'
                  ) : (
                    <>
                      <Plus size={16} />
                      Create Batch
                    </>
                  )}
                </button>
              </div>
            </div>
          )}
        </form>
      </div>
    </div>
  );
}

export default CreateBatch;
