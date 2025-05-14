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

      <div className={cx('left-side', 'w-1/2')}>
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
            First time here? <Link href={'register'}>Create one</Link>
          </div>
        </div>
      </div>

      <div className={cx('right-side', 'w-1/2')}></div>
    </div>
  );
}
export default Login;
