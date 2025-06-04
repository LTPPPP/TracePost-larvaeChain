'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

interface UserAnalyticsData {
  success: boolean;
  message: string;
  data: {
    active_users_by_role: {
      admin: number;
      distributor: number;
      hatchery: number;
      user: number;
    };
    login_frequency: {
      last_30_days: number;
      last_7_days: number;
      today: number;
      yesterday: number;
    };
    api_endpoint_usage: Record<string, number>;
    most_active_users: Array<{
      user_id: number;
      username: string;
      request_count: number;
      last_active: string;
    }>;
    user_growth: {
      last_30_days: number;
      last_7_days: number;
      last_90_days: number;
      today: number;
      yesterday: number;
    };
    last_updated: string;
  };
}

export const RoleInfo = () => {
  const [roles, setRoles] = useState<UserAnalyticsData['data']['active_users_by_role'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchUserAnalytics = async () => {
      try {
        setLoading(true);
        setError(null);

        // Retrieve token from localStorage or sessionStorage
        const token = localStorage.getItem('token') || sessionStorage.getItem('token');

        if (!token) {
          setError('Authentication token is missing. Please log in.');
          router.push('/login');
          return;
        }

        const response = await fetch('http://localhost:8080/api/v1/admin/analytics/users', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`
          }
        });

        if (!response.ok) {
          if (response.status === 401) {
            setError('Unauthorized: Invalid or expired token. Please log in again.');
            localStorage.removeItem('token');
            sessionStorage.removeItem('token');
            router.push('/login');
            return;
          }
          throw new Error(`Failed to fetch user analytics: ${response.statusText}`);
        }

        const data: UserAnalyticsData = await response.json();
        if (!data.success) {
          throw new Error(data.message || 'Failed to retrieve user analytics data');
        }

        setRoles(data.data.active_users_by_role);
      } catch (err) {
        console.error('Error fetching user analytics:', err);
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchUserAnalytics();
  }, [router]);

  if (loading) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>Loading role info...</div>;
  }

  if (error) {
    return <div className='text-center py-4 text-red-500 dark:text-red-400'>Error: {error}</div>;
  }

  if (!roles) {
    return <div className='text-center py-4 text-gray-500 dark:text-gray-400'>No role data available.</div>;
  }

  return (
    <div className='p-6'>
      <h2 className='text-xl font-bold text-gray-800 dark:text-white/90 mb-4'>Active Users by Role</h2>
      <div className='bg-white dark:bg-white/[0.03] rounded-xl border border-gray-200 dark:border-gray-800 p-5'>
        <p className='text-gray-700 dark:text-gray-300'>
          <span className='font-semibold'>Admins:</span> {(roles.admin ?? 0).toLocaleString()}
        </p>
        <p className='text-gray-700 dark:text-gray-300'>
          <span className='font-semibold'>Distributors:</span> {(roles.distributor ?? 0).toLocaleString()}
        </p>
        <p className='text-gray-700 dark:text-gray-300'>
          <span className='font-semibold'>Hatcheries:</span> {(roles.hatchery ?? 0).toLocaleString()}
        </p>
        <p className='text-gray-700 dark:text-gray-300'>
          <span className='font-semibold'>Users:</span> {(roles.user ?? 0).toLocaleString()}
        </p>
      </div>
    </div>
  );
};
