'use client';
import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useModal } from '../../hooks/useModal';
import { Modal } from '../ui/modal';
import Button from '../ui/button/Button';
import Input from '../form/input/InputField';
import Label from '../form/Label';

interface UserData {
  data: {
    avatar_url: string;
    company: Record<string, unknown>;
    company_id: number;
    created_at: string;
    date_of_birth: string;
    email: string;
    full_name: string;
    id: number;
    is_active: boolean;
    last_login: string;
    phone: string;
    role: string;
    updated_at: string;
    username: string;
  };
  message: string;
  success: boolean;
}

export default function UserInfoCard() {
  const { isOpen, openModal, closeModal } = useModal();
  const [user, setUser] = useState<UserData['data'] | null>(null);
  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    phone: '',
    role: ''
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchUserData = async () => {
      try {
        setLoading(true);
        setError(null);

        const token = localStorage.getItem('token') || sessionStorage.getItem('token');

        if (!token) {
          setError('Authentication token is missing. Please log in.');
          router.push('/login');
          return;
        }

        const response = await fetch('http://localhost:8080/api/v1/users/me', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`
          }
        });

        const data: UserData = await response.json();
        // in data object, check if success is true
        console.log('User data fetched:', data);
        if (!data.success) {
          throw new Error(data.message || 'Failed to retrieve user data');
        }

        const userData = data.data;
        setUser(userData);
        // Split full_name into first and last name (assuming space-separated)
        const [firstName, lastName] = userData.full_name.split(' ') || [userData.full_name, ''];
        setFormData({
          full_name: `${firstName} ${lastName}`.trim(),
          email: userData.email,
          phone: userData.phone || '',
          role: userData.role
        });
      } catch (err) {
        console.error('Error fetching user data:', err);
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchUserData();
  }, [router]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSave = () => {
    // Placeholder for save logic (replace with API call)
    console.log('Saving changes:', formData);
    // Example API call:
    // fetch('http://localhost:8080/api/v1/users/me', {
    //   method: 'PATCH',
    //   headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    //   body: JSON.stringify(formData),
    // });
    closeModal();
  };

  if (loading) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>Loading user info...</div>;
  }

  if (error) {
    return <div className='text-center py-4 text-red-500 dark:text-red-400'>Error: {error}</div>;
  }

  if (!user) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>No user data available.</div>;
  }

  return (
    <div className='p-5 border border-gray-200 rounded-2xl dark:border-gray-800 lg:p-6'>
      <div className='flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between'>
        <div>
          <h4 className='text-lg font-semibold text-gray-800 dark:text-white/90 lg:mb-6'>Personal Information</h4>
          <div className='grid grid-cols-1 gap-4 lg:grid-cols-2 lg:gap-7 2xl:gap-x-32'>
            <div>
              <p className='mb-2 text-xs leading-normal text-gray-500 dark:text-gray-400'>First Name</p>
              <p className='text-sm font-medium text-gray-800 dark:text-white/90'>
                {user.full_name.split(' ')[0] || ''}
              </p>
            </div>
            <div>
              <p className='mb-2 text-xs leading-normal text-gray-500 dark:text-gray-400'>Last Name</p>
              <p className='text-sm font-medium text-gray-800 dark:text-white/90'>
                {user.full_name.split(' ').slice(1).join(' ') || ''}
              </p>
            </div>
            <div>
              <p className='mb-2 text-xs leading-normal text-gray-500 dark:text-gray-400'>Email address</p>
              <p className='text-sm font-medium text-gray-800 dark:text-white/90'>{user.email}</p>
            </div>
            <div>
              <p className='mb-2 text-xs leading-normal text-gray-500 dark:text-gray-400'>Phone</p>
              <p className='text-sm font-medium text-gray-800 dark:text-white/90'>{user.phone || 'N/A'}</p>
            </div>
            <div>
              <p className='mb-2 text-xs leading-normal text-gray-500 dark:text-gray-400'>Bio</p>
              <p className='text-sm font-medium text-gray-800 dark:text-white/90'>{user.role}</p>
            </div>
          </div>
        </div>
        <button
          onClick={openModal}
          className='flex w-full items-center justify-center gap-2 rounded-full border border-gray-300 bg-white px-4 py-3 text-sm font-medium text-gray-700 shadow-theme-xs hover:bg-gray-50 hover:text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-white/[0.03] dark:hover:text-gray-200 lg:inline-flex lg:w-auto'
        >
          <svg
            className='fill-current'
            width='18'
            height='18'
            viewBox='0 0 18 18'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
          >
            <path
              fillRule='evenodd'
              clipRule='evenodd'
              d='M15.0911 2.78206C14.2125 1.90338 12.7878 1.90338 11.9092 2.78206L4.57524 10.116C4.26682 10.4244 4.0547 10.8158 3.96468 11.2426L3.31231 14.3352C3.25997 14.5833 3.33653 14.841 3.51583 15.0203C3.69512 15.1996 3.95286 15.2761 4.20096 15.2238L7.29355 14.5714C7.72031 14.4814 8.11172 14.2693 8.42013 13.9609L15.7541 6.62695C16.6327 5.74827 16.6327 4.32365 15.7541 3.44497L15.0911 2.78206ZM12.9698 3.84272C13.2627 3.54982 13.7376 3.54982 14.0305 3.84272L14.6934 4.50563C14.9863 4.79852 14.9863 5.2734 14.6934 5.56629L14.044 6.21573L12.3204 4.49215L12.9698 3.84272ZM11.2597 5.55281L5.6359 11.1766C5.53309 11.2794 5.46238 11.4099 5.43238 11.5522L5.01758 13.5185L6.98394 13.1037C7.1262 13.0737 7.25666 13.003 7.35947 12.9002L12.9833 7.27639L11.2597 5.55281Z'
              fill=''
            />
          </svg>
          Edit
        </button>
      </div>

      <Modal isOpen={isOpen} onClose={closeModal} className='max-w-[700px] m-4'>
        <div className='no-scrollbar relative w-full max-w-[700px] overflow-y-auto rounded-3xl bg-white p-4 dark:bg-gray-900 lg:p-11'>
          <div className='px-2 pr-14'>
            <h4 className='mb-2 text-2xl font-semibold text-gray-800 dark:text-white/90'>Edit Personal Information</h4>
            <p className='mb-6 text-sm text-gray-500 dark:text-gray-400 lg:mb-7'>
              Update your details to keep your profile up-to-date.
            </p>
          </div>
          <form className='flex flex-col'>
            <div className='custom-scrollbar h-[450px] overflow-y-auto px-2 pb-3'>
              <div className='mt-7'>
                <h5 className='mb-5 text-lg font-medium text-gray-800 dark:text-white/90 lg:mb-6'>
                  Personal Information
                </h5>
                <div className='grid grid-cols-1 gap-x-6 gap-y-5 lg:grid-cols-2'>
                  <div className='col-span-2 lg:col-span-1'>
                    <Label>First Name</Label>
                    <Input
                      type='text'
                      name='firstName'
                      defaultValue={formData.full_name.split(' ')[0] || ''}
                      onChange={handleChange}
                      placeholder='Enter first name'
                    />
                  </div>
                  <div className='col-span-2 lg:col-span-1'>
                    <Label>Last Name</Label>
                    <Input
                      type='text'
                      name='lastName'
                      defaultValue={formData.full_name.split(' ').slice(1).join(' ') || ''}
                      onChange={(e) => {
                        const [firstName] = formData.full_name.split(' ');
                        setFormData((prev) => ({ ...prev, full_name: `${firstName} ${e.target.value}`.trim() }));
                      }}
                      placeholder='Enter last name'
                    />
                  </div>
                  <div className='col-span-2 lg:col-span-1'>
                    <Label>Email Address</Label>
                    <Input
                      type='email'
                      name='email'
                      defaultValue={formData.email}
                      onChange={handleChange}
                      placeholder='Enter email address'
                    />
                  </div>
                  <div className='col-span-2 lg:col-span-1'>
                    <Label>Phone</Label>
                    <Input
                      type='text'
                      name='phone'
                      defaultValue={formData.phone}
                      onChange={handleChange}
                      placeholder='Enter phone number'
                    />
                  </div>
                  <div className='col-span-2'>
                    <Label>Bio</Label>
                    <Input
                      type='text'
                      name='role'
                      defaultValue={formData.role}
                      onChange={handleChange}
                      placeholder='Enter bio'
                    />
                  </div>
                </div>
              </div>
            </div>
            <div className='flex items-center gap-3 px-2 mt-6 lg:justify-end'>
              <Button size='sm' variant='outline' onClick={closeModal}>
                Close
              </Button>
              <Button size='sm' onClick={handleSave}>
                Save Changes
              </Button>
            </div>
          </form>
        </div>
      </Modal>
    </div>
  );
}
