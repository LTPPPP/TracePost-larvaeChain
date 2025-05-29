'use client';
import Link from 'next/link';
import Image from 'next/image';
import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import styles from './Login.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

//

interface LoginRequest {
  username: string;
  password: string;
}

// Interface cho login response
interface LoginResponseData {
  access_token: string;
  token_type: string;
  expires_in: number;
}

interface LoginResponse {
  success: boolean;
  message: string;
  data: LoginResponseData;
}
//

function Login() {
  const [formData, setFormData] = useState<LoginRequest>({
    username: '',
    password: ''
  });
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const router = useRouter();

  // Handle input changes
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value
    }));
    // Clear error when user starts typing
    if (error) setError(null);
  };

  // Handle show password toggle
  const handleShowPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setShowPassword(e.target.checked);
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    // Validation
    if (!formData.username.trim() || !formData.password.trim()) {
      setError('Please enter both username and password');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });

      console.log('Response status:', response.status); // Debug

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Invalid username or password');
        } else if (response.status === 400) {
          throw new Error('Please check your input');
        } else {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Login failed. Please try again.');
        }
      }

      const data: LoginResponse = await response.json();
      console.log('Parsed response data:', data); // Debug

      // Save token to localStorage
      try {
        const accessToken = data.data.access_token;
        const tokenType = data.data.token_type;
        const expiresIn = data.data.expires_in;

        console.log('Saving token:', accessToken); // Debug

        localStorage.setItem('token', accessToken);
        localStorage.setItem('tokenType', tokenType);
        localStorage.setItem('tokenExpires', (Date.now() + expiresIn * 1000).toString());

        const savedToken = localStorage.getItem('token');
        console.log('Token saved to localStorage:', savedToken);
        if (!savedToken) {
          throw new Error('Token not saved to localStorage');
        }
      } catch (storageError) {
        console.error('Failed to save to localStorage:', storageError);
        setError('Unable to save authentication data. Please check your browser settings.');
        return;
      }
      // Redirect to dashboard
      router.push('/admin');
    } catch (err) {
      console.error('Login error:', err);
      setError(err instanceof Error ? err.message : 'An unexpected error occurred');
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
              <div className={cx('left-title')}>Log In</div>
              <div className={cx('left-description')}>To your shrimp traceability network</div>

              <div className={cx('left-subdescription')}>
                Access verified data, track every batch, and stay compliant with international seafood standards.
              </div>
            </div>
          </div>

          <div className={cx('action')}>
            First time here? <Link href={'/register'}>Create one</Link>
          </div>
        </div>
      </div>

      <div className={cx('right-side', 'w-full', 'lg:w-1/2')}>
        <div className={cx('right-container')}>
          <h2 className={cx('form-title')}>Sign In</h2>

          <form className={cx('login-form')} onSubmit={handleSubmit}>
            {/* Error message */}
            {error && (
              <div
                className={cx('error-message')}
                style={{
                  padding: '12px',
                  marginBottom: '16px',
                  backgroundColor: '#fee2e2',
                  border: '1px solid #fecaca',
                  color: '#dc2626',
                  borderRadius: '6px',
                  fontSize: '14px'
                }}
              >
                {error}
              </div>
            )}
            <div className={cx('form-group')}>
              <label htmlFor='email' className={cx('form-label')}>
                Email
              </label>
              <input
                type='text'
                id='username'
                name='username'
                className={cx('form-input')}
                placeholder='Username'
                value={formData.username}
                onChange={handleInputChange}
                disabled={loading}
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
                disabled={loading}
                required
              />
              <div className={cx('show-password')}>
                <input
                  type='checkbox'
                  id='showPassword'
                  className={cx('checkbox')}
                  checked={showPassword}
                  onChange={handleShowPasswordChange}
                  disabled={loading}
                />
                <label htmlFor='showPassword' className={cx('checkbox-label')}>
                  Show password
                </label>
              </div>
            </div>

            <button
              type='submit'
              className={cx('sign-in-btn')}
              disabled={loading}
              style={{
                opacity: loading ? 0.7 : 1,
                cursor: loading ? 'not-allowed' : 'pointer'
              }}
            >
              {loading ? 'SIGNING IN...' : 'SIGN IN'}
            </button>

            <div className={cx('divider')}>
              <span>or</span>
            </div>

            <button type='button' className={cx('google-btn')}>
              <Image src='/img/auth/google-icon.png' alt='Google' width={24} height={24} />
              Continue with Google
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
export default Login;
