'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getListCompany, ApiCompany, ApiResponse } from '@/api/company';
import { register, RegisterData } from '@/api/auth';

import styles from './Register.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

interface RegisterFormData {
  email: string;
  password: string;
  username: string;
  company_id: string;
  role: string;
}

const ROLES = [
  { value: 'user', label: 'User' },
  { value: 'hatchery', label: 'Hatchery' },
  { value: 'distributor', label: 'Distributor' }
];

function Register() {
  const router = useRouter();
  const [formData, setFormData] = useState<RegisterFormData>({
    email: '',
    password: '',
    username: '',
    company_id: '',
    role: ''
  });

  const [companies, setCompanies] = useState<ApiCompany[]>([]);
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    const fetchCompanies = async () => {
      try {
        const response: ApiResponse<ApiCompany> = await getListCompany();
        if (response.success) {
          setCompanies(response.data);
        }
      } catch (error) {
        console.error('Error loading companies:', error);
        setError('Failed to load companies');
      }
    };

    fetchCompanies();
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // Validation
    if (!formData.email || !formData.password || !formData.username || !formData.company_id || !formData.role) {
      setError('Please fill in all fields');
      return;
    }

    setLoading(true);

    try {
      const registerData: RegisterData = {
        email: formData.email,
        password: formData.password,
        username: formData.username,
        company_id: formData.company_id,
        role: formData.role
      };

      const response = await register(registerData);
      const result = await response.json();

      if (response.ok) {
        setSuccess('Registration successful!');

        // Reset
        setFormData({
          email: '',
          password: '',
          username: '',
          company_id: '',
          role: ''
        });

        // Redirect to login
        setTimeout(() => {
          router.push('/login');
        }, 1000);
      } else {
        if (response.status === 409) {
          setError('Email already exists');
        } else {
          setError(result.message || 'Registration failed');
        }
      }
    } catch (error) {
      console.error('Registration error:', error);
      setError('An error occurred during registration');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={cx('wrapper')}>
      <Image
        className={cx('float-circle')}
        src={'/img/auth/float_circle.png'}
        alt='float-circle'
        width={150}
        height={150}
      />

      <div className={cx('left-side', 'w-full', 'lg:w-1/2')}>
        <Link className={cx('logo')} href={'/'}>
          <Image src={'/img/logo.png'} alt='logo' width={50} height={50} />
        </Link>

        <div className={cx('left-container')}>
          <div className={cx('left-content')}>
            <div className={cx('left-slogan')}>
              TRACKTO<span>TRUTH</span>
            </div>

            <div>
              <div className={cx('left-title')}>Register</div>
              <div className={cx('left-description')}>To enable global traceability</div>

              <div className={cx('left-subdescription')}>
                Become part of a next-gen solution using blockchain, smart contracts, and IPFS to meet international
                standards and elevate your brand.
              </div>
            </div>
          </div>

          <div className={cx('action')}>
            Already a member? <Link href={'/login'}>Log in</Link>
          </div>
        </div>
      </div>

      <div className={cx('right-side', 'w-full', 'lg:w-1/2')}>
        <div className={cx('right-container')}>
          <h2 className={cx('form-title')}>Sign Up</h2>

          <form className={cx('register-form')} onSubmit={handleSubmit}>
            {error && <div className={cx('error-message')}>{error}</div>}

            {success && <div className={cx('success-message')}>{success}</div>}

            <div className={cx('form-group')}>
              <label htmlFor='username' className={cx('form-label')}>
                Username
              </label>
              <input
                type='text'
                id='username'
                name='username'
                className={cx('form-input')}
                placeholder='Username'
                value={formData.username}
                onChange={handleInputChange}
                required
              />
            </div>

            <div className={cx('form-group')}>
              <label htmlFor='email' className={cx('form-label')}>
                Email
              </label>
              <input
                type='email'
                id='email'
                name='email'
                className={cx('form-input')}
                placeholder='Email'
                value={formData.email}
                onChange={handleInputChange}
                required
              />
            </div>

            <div className={cx('form-group')}>
              <label htmlFor='password' className={cx('form-label')}>
                Password
              </label>
              <input
                type={showPassword ? 'text' : 'password'}
                id='password'
                name='password'
                className={cx('form-input')}
                placeholder='Password'
                value={formData.password}
                onChange={handleInputChange}
                required
              />
              <div className={cx('show-password')}>
                <input
                  type='checkbox'
                  id='showPassword'
                  className={cx('checkbox')}
                  checked={showPassword}
                  onChange={(e) => setShowPassword(e.target.checked)}
                />
                <label htmlFor='showPassword' className={cx('checkbox-label')}>
                  Show password
                </label>
              </div>
            </div>

            <div className={cx('form-group')}>
              <label htmlFor='company_id' className={cx('form-label')}>
                Company
              </label>
              <select
                id='company_id'
                name='company_id'
                className={cx('form-input')}
                value={formData.company_id}
                onChange={handleInputChange}
                required
              >
                <option value=''>Select a company</option>
                {companies.map((company) => (
                  <option key={company.id} value={company.id.toString()}>
                    {company.name}
                  </option>
                ))}
              </select>
            </div>

            <div className={cx('form-group')}>
              <label htmlFor='role' className={cx('form-label')}>
                Role
              </label>
              <select
                id='role'
                name='role'
                className={cx('form-input')}
                value={formData.role}
                onChange={handleInputChange}
                required
              >
                <option value=''>Select a role</option>
                {ROLES.map((role) => (
                  <option key={role.value} value={role.value}>
                    {role.label}
                  </option>
                ))}
              </select>
            </div>

            <button type='submit' className={cx('sign-up-btn')} disabled={loading}>
              {loading ? 'Creating Account...' : 'Sign Up'}
            </button>

            {/* <div className={cx('divider')}>
              <span>or</span>
            </div>

            <button type='button' className={cx('google-btn')}>
              <Image src='/img/auth/google-icon.png' alt='Google' width={24} height={24} />
              Continue with Google
            </button> */}
          </form>
        </div>
      </div>
    </div>
  );
}

export default Register;
