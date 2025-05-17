import Link from 'next/link';
import Image from 'next/image';

import styles from './Register.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

function Register() {
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

          <form className={cx('register-form')}>
            <div className={cx('form-group')}>
              <label htmlFor='email' className={cx('form-label')}>
                Email
              </label>
              <input type='email' id='email' className={cx('form-input')} placeholder='Email' />
            </div>

            <div className={cx('form-group')}>
              <label htmlFor='password' className={cx('form-label')}>
                Password
              </label>
              <input type='password' id='password' className={cx('form-input')} placeholder='Password' />
              <div className={cx('show-password')}>
                <input type='checkbox' id='showPassword' className={cx('checkbox')} />
                <label htmlFor='showPassword' className={cx('checkbox-label')}>
                  Show password
                </label>
              </div>
            </div>

            <button type='submit' className={cx('sign-up-btn')}>
              Sign Up
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
export default Register;
