'use client';
import { useState, useEffect } from 'react';
import { getProfile } from '@/api/profile';
import { getUserInfo, clearAuthData } from '@/utils/auth';

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
  company: any;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login: string;
}

export const useProfile = () => {
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
        setProfile(null);
        setLoading(false);
        return;
      }

      const response = await getProfile();
      setProfile(response.data);
    } catch (err) {
      console.error('Error fetching profile:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch profile');

      if (err instanceof Error && err.message.includes('token')) {
        clearAuthData();
        setProfile(null);
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
