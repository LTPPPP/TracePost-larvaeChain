import Image from 'next/image';
import styles from './Home.module.scss';
import classNames from 'classnames/bind';
import Link from 'next/link';
const cx = classNames.bind(styles);

function Home() {
  return (
    <div className={cx('wrapper')}>
      {/* NAV */}
      <div className={cx('nav')}>
        <div className={cx('logo')}>
          <Link href={'/'}>
            <Image src={'/img/logo_word.png'} alt='logo' height={75} width={300} />
          </Link>
        </div>

        <div className={cx('nav-list')}>
          <div className={cx('nav-item')}>
            <Link href={'#home'}>Home</Link>
          </div>

          <div className={cx('nav-item')}>
            <Image src={'/img/home/star.png'} alt='star' height={56} width={56} />
            <Link href={'#features'}>Features</Link>
          </div>

          <div className={cx('nav-item')}>
            <Link href={'#enterprise'}>Enterprise</Link>
          </div>
        </div>
      </div>

      {/* HERO */}
      <section className={cx('hero')} id='home'>
        <div className={cx('hero-content')}>
          <div className={cx('hero-container')}>
            <div className={cx('hero-title')}>We are</div>
            <Image src={'/img/home/star.png'} alt='star' height={56} width={56} />
            <div className={cx('hero-subtitle')}>supporting</div>
          </div>

          <div className={cx('type-text')}>
            Trans<span>parency</span>
          </div>

          <div className={cx('hero-description')}>
            A blockchain-based platform ensuring trust, traceability, and international compliance in shrimp hatchery
            supply chains.
          </div>

          <Link href={'/login'} className={cx('hero-action')}>
            Get Started
          </Link>
        </div>
      </section>
    </div>
  );
}
export default Home;
