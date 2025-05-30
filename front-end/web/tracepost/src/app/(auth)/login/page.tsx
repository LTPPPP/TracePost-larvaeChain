'use client';
import React, { useState } from 'react';

import Link from 'next/link';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { login } from '@/api/auth';

import styles from './Login.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

const roleRedirectMap: Record<string, string> = {
  admin: '/workspace',
  user: '/company-list',
  distributor: '/distributor',
  hatchery: '/workspace'
};

function Login() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    if (name === 'username') setUsername(value);
    else if (name === 'password') setPassword(value);
  };

  const handleShowPasswordChange = () => {
    setShowPassword(!showPassword);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await login(username, password);
      const result = await response.json();
      console.log(result.data);

      if (!response.ok) {
        throw new Error(result.error || 'Login failed');
      }

      // REDICRECT
      const role = result.data?.role;

      const redirectPath = roleRedirectMap[role];
      router.push(redirectPath);
    } catch (err: unknown) {
      console.log(err);
      setError(err instanceof Error ? err.message : 'Login failed');
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
            {error && <div className={cx('error-message')}>{error}</div>}
            <div className={cx('form-group')}>
              <label htmlFor='username' className={cx('form-label')}>
                Email
              </label>
              <input
                type='text'
                id='username'
                name='username'
                className={cx('form-input')}
                placeholder='Username'
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
