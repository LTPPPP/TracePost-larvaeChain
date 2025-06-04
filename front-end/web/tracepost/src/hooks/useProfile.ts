'use client';
import { useState, useEffect } from 'react';
import { getProfile } from '@/api/profile';
import { getUserInfo, clearAuthData } from '@/utils/auth';
import { useRouter } from 'next/navigation';

interface Company {
  id: number;
  name: string;
  location?: string;
  contact_info?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

interface Profile {
  id: number;
  username: string;
  full_name: string;
  email: string;
  phone: string;
  date_of_birth: string;
  avatar_url: string;
  role: string;
  company_id: number;
  company?: Company;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login: string;
}

export const useProfile = () => {
  const router = useRouter();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = async () => {
    try {
      setLoading(true);
      setError(null);

      // Check if user is logged in
      const userInfo = getUserInfo();
      if (!userInfo) {
        console.log('No user info found, redirecting to login');
        setProfile(null);
        setLoading(false);
        router.push('/login');
        return;
      }

      console.log('Fetching profile for user:', userInfo.user_id);
      const response = await getProfile();
      console.log('Profile response:', response);

      setProfile(response.data);
    } catch (err) {
      console.error('Error fetching profile:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch profile';
      setError(errorMessage);

      // Error
      if (
        errorMessage.includes('authentication') ||
        errorMessage.includes('token') ||
        errorMessage.includes('401') ||
        errorMessage.includes('403')
      ) {
        console.log('Auth error detected, clearing data and redirecting');
        clearAuthData();
        setProfile(null);
        router.push('/login');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  return {
    profile,
    loading,
    error,
    refetch: fetchProfile
  };
};
