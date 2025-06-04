'use client';

import { useState } from 'react';
import { Thermometer, Droplet, Waves, Clock, Shrimp, Sigma, Edit3, X, Save, Plus } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './HatcheryCard.module.scss';

// Import API functions from centralized batch.ts
import {
  createEnvironment,
  updateEnvironment,
  CreateEnvironmentData,
  UpdateEnvironmentData
  // handleApiError
} from '@/api/batch';

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

interface HatcheryCardProps {
  data: HatcheryData[];
  onUpdateEnvironment?: (id: string, updatedData: Partial<HatcheryData>) => void;
}

interface EditFormData {
  temperature: number;
  ph: number;
  salinity: number;
  density: number;
  age: number;
}

function HatcheryCard({ data, onUpdateEnvironment }: HatcheryCardProps) {
  const [selectedPond, setSelectedPond] = useState<HatcheryData | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editForm, setEditForm] = useState<EditFormData>({
    temperature: 0,
    ph: 0,
    salinity: 0,
    density: 0,
    age: 0
  });
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateMode, setIsCreateMode] = useState(false);

  // Khoảng ổn định
  const TEMP_MIN = 28;
  const TEMP_MAX = 30;
  const PH_MIN = 7.5;
  const PH_MAX = 8.5;

  /**
   * LOW   - BLUE
   * STABLE- GREEN
   * HIGH  - RED
   */
  const getTemperatureColor = (temp: number) => {
    if (temp < TEMP_MIN) return 'blue';
    if (temp > TEMP_MAX) return 'red';
    return 'green';
  };

  const getPhColor = (ph: number) => {
    if (ph < PH_MIN) return 'blue';
    if (ph > PH_MAX) return 'red';
    return 'green';
  };

  const hasEnvironmentData = (pond: HatcheryData) => {
    return pond.temperature > 0 || pond.ph > 0 || pond.salinity > 0 || pond.density > 0 || pond.age > 0;
  };

  const handleCardClick = (pond: HatcheryData) => {
    const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}');
    if (userInfo.role !== 'hatchery') {
      return;
    }

    setSelectedPond(pond);
    const hasData = hasEnvironmentData(pond);
    setIsCreateMode(!hasData);

    setEditForm({
      temperature: hasData ? pond.temperature : 0,
      ph: hasData ? pond.ph : 0,
      salinity: hasData ? pond.salinity : 0,
      density: hasData ? pond.density : 0,
      age: hasData ? pond.age : 0
    });
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedPond(null);
    setIsCreateMode(false);
  };

  const handleInputChange = (field: keyof EditFormData, value: string) => {
    const numericValue = parseFloat(value) || 0;
    setEditForm((prev) => ({
      ...prev,
      [field]: numericValue
    }));
  };

  const handleSave = async () => {
    if (!selectedPond) return;

    setIsLoading(true);
    try {
      if (isCreateMode) {
        // Create new environment
        const createData: CreateEnvironmentData = {
          age: editForm.age,
          batch_id: selectedPond.batchId || 0,
          density: editForm.density,
          ph: editForm.ph,
          salinity: editForm.salinity,
          temperature: editForm.temperature
        };

        const response = await createEnvironment(createData);

        if (response.success) {
          if (onUpdateEnvironment) {
            onUpdateEnvironment(selectedPond.id, editForm);
          }
          handleCloseModal();
        } else {
          throw new Error(response.message || 'Failed to create environment');
        }
      } else {
        // Update existing environment
        const updateData: UpdateEnvironmentData = {
          temperature: editForm.temperature,
          ph: editForm.ph,
          salinity: editForm.salinity,
          density: editForm.density,
          age: editForm.age
        };

        const response = await updateEnvironment(selectedPond.id, updateData);

        if (response.success) {
          if (onUpdateEnvironment) {
            onUpdateEnvironment(selectedPond.id, editForm);
          }
          handleCloseModal();
        } else {
          throw new Error(response.message || 'Failed to update environment');
        }
      }
    } catch (error) {
      console.error(`Error ${isCreateMode ? 'creating' : 'updating'} environment:`, error);
      // const errorMessage = handleApiError(error);
    } finally {
      setIsLoading(false);
    }
  };

  const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}');
  const isHatcheryUser = userInfo.role === 'hatchery';

  return (
    <>
      <div className={cx('environment-container')}>
        <div className={cx('section-header')}>
          <h2 className={cx('section-title')}>Active Batches ({data.length})</h2>
        </div>

        <div className={cx('ponds-grid')}>
          {data.map((pond) => {
            const hasData = hasEnvironmentData(pond);

            return (
              <div
                key={pond.id}
                className={cx('pond-card', { clickable: isHatcheryUser })}
                onClick={() => handleCardClick(pond)}
              >
                {isHatcheryUser && (
                  <div className={cx('edit-icon', { 'create-mode': !hasData })}>
                    {hasData ? <Edit3 size={20} /> : <Plus size={20} />}
                  </div>
                )}

                <div className={cx('pond-header')}>
                  <div className={cx('pond-title')}>
                    <h3 className={cx('pond-name')}>
                      {pond.name} - {pond.species}
                    </h3>
                  </div>
                </div>

                <div className={cx('pond-details')}>
                  <div className={cx('detail-item')}>
                    <Sigma className={cx('detail-icon')} size={25} />
                    <span className={cx('detail-label')}>Quantity:</span>
                    <span className={cx('detail-value')}>{pond.quantity.toLocaleString()}</span>
                  </div>

                  <div className={cx('detail-item')}>
                    <Thermometer className={cx('detail-icon', getTemperatureColor(pond.temperature))} size={25} />
                    <span className={cx('detail-label')}>Temp:</span>
                    <span className={cx('detail-value', getTemperatureColor(pond.temperature))}>
                      {pond.temperature > 0 ? `${pond.temperature}°C` : 'N/A'}
                    </span>
                  </div>

                  <div className={cx('detail-item')}>
                    <Droplet className={cx('detail-icon', getPhColor(pond.ph))} size={25} />
                    <span className={cx('detail-label')}>pH:</span>
                    <span className={cx('detail-value', getPhColor(pond.ph))}>{pond.ph > 0 ? pond.ph : 'N/A'}</span>
                  </div>

                  <div className={cx('detail-item')}>
                    <Waves className={cx('detail-icon')} size={25} />
                    <span className={cx('detail-label')}>Salinity:</span>
                    <span className={cx('detail-value')}>{pond.salinity > 0 ? `${pond.salinity} ppt` : 'N/A'}</span>
                  </div>

                  <div className={cx('detail-item')}>
                    <Shrimp className={cx('detail-icon')} size={25} />
                    <span className={cx('detail-label')}>Density:</span>
                    <span className={cx('detail-value')}>{pond.density > 0 ? `${pond.density}/m³` : 'N/A'}</span>
                  </div>

                  <div className={cx('detail-item')}>
                    <Clock className={cx('detail-icon')} size={25} />
                    <span className={cx('detail-label')}>Age:</span>
                    <span className={cx('detail-value')}>{pond.age > 0 ? `${pond.age} days` : 'N/A'}</span>
                  </div>
                </div>

                {!hasData && isHatcheryUser && (
                  <div className={cx('no-data-indicator')}>
                    <span>Click to add environment data</span>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Edit/Create Modal */}
      {isModalOpen && selectedPond && (
        <div className={cx('modal-overlay')}>
          <div className={cx('modal-content')}>
            <div className={cx('modal-header')}>
              <h3 className={cx('modal-title')}>
                {isCreateMode ? 'Add Environment Data' : 'Edit Environment'} - {selectedPond.name}
              </h3>
              <button className={cx('close-button')} onClick={handleCloseModal} disabled={isLoading}>
                <X size={24} />
              </button>
            </div>

            <div className={cx('modal-body')}>
              <div className={cx('form-grid')}>
                <div className={cx('form-group')}>
                  <label className={cx('form-label')}>
                    <Thermometer size={20} />
                    Temperature (°C)
                  </label>
                  <input
                    type='number'
                    step='0.1'
                    value={editForm.temperature}
                    onChange={(e) => handleInputChange('temperature', e.target.value)}
                    className={cx('form-input', getTemperatureColor(editForm.temperature))}
                    disabled={isLoading}
                  />
                  <span className={cx('form-hint')}>
                    Optimal: {TEMP_MIN}°C - {TEMP_MAX}°C
                  </span>
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('form-label')}>
                    <Droplet size={20} />
                    pH Level
                  </label>
                  <input
                    type='number'
                    step='0.1'
                    value={editForm.ph}
                    onChange={(e) => handleInputChange('ph', e.target.value)}
                    className={cx('form-input', getPhColor(editForm.ph))}
                    disabled={isLoading}
                  />
                  <span className={cx('form-hint')}>
                    Optimal: {PH_MIN} - {PH_MAX}
                  </span>
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('form-label')}>
                    <Waves size={20} />
                    Salinity (ppt)
                  </label>
                  <input
                    type='number'
                    step='0.1'
                    value={editForm.salinity}
                    onChange={(e) => handleInputChange('salinity', e.target.value)}
                    className={cx('form-input')}
                    disabled={isLoading}
                  />
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('form-label')}>
                    <Shrimp size={20} />
                    Density (/m³)
                  </label>
                  <input
                    type='number'
                    step='1'
                    value={editForm.density}
                    onChange={(e) => handleInputChange('density', e.target.value)}
                    className={cx('form-input')}
                    disabled={isLoading}
                  />
                </div>

                <div className={cx('form-group')}>
                  <label className={cx('form-label')}>
                    <Clock size={20} />
                    Age (days)
                  </label>
                  <input
                    type='number'
                    step='1'
                    value={editForm.age}
                    onChange={(e) => handleInputChange('age', e.target.value)}
                    className={cx('form-input')}
                    disabled={isLoading}
                  />
                </div>
              </div>
            </div>

            <div className={cx('modal-footer')}>
              <button className={cx('cancel-button')} onClick={handleCloseModal} disabled={isLoading}>
                Cancel
              </button>
              <button className={cx('save-button')} onClick={handleSave} disabled={isLoading}>
                <Save size={18} />
                {isLoading
                  ? isCreateMode
                    ? 'Creating...'
                    : 'Saving...'
                  : isCreateMode
                  ? 'Create Environment'
                  : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

export default HatcheryCard;
