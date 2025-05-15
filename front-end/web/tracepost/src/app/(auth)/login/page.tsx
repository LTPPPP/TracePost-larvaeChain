import Link from 'next/link';
import Image from 'next/image';

import styles from './Login.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

function Login() {
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

          <form className={cx('login-form')}>
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

            <button type='submit' className={cx('sign-in-btn')}>
              SIGN IN
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
