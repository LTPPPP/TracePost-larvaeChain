'use client';

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import NotificationDropdown from '@/components/header/NotificationDropdown';
import UserDropdown from '@/components/header/UserDropdown';

import { FileText, BookOpen, Smartphone, MoveRight, YoutubeIcon, GithubIcon, Mail } from 'lucide-react';

import styles from '../distribute.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

type ImagePosition = Partial<Record<'top' | 'bottom' | 'left' | 'right' | 'transform', string>>;

interface ImageItem {
  id: number;
  src: string;
  alt: string;
  position: ImagePosition;
}

type GuideKey = 'document' | 'guide' | 'app';

interface DivWithHandlers extends HTMLDivElement {
  _moveHandler?: (e: MouseEvent) => void;
  _leaveHandler?: () => void;
}

function Home() {
  // GUIDE
  const [activeGuide, setActiveGuide] = useState<GuideKey>('document');

  const guideRefs: Record<GuideKey, React.RefObject<HTMLDivElement | null>> = {
    document: useRef<HTMLDivElement>(null),
    guide: useRef<HTMLDivElement>(null),
    app: useRef<HTMLDivElement>(null)
  };

  // ACHIVIE
  const itemsRef = useRef<(HTMLDivElement | null)[]>([]);

  const images: ImageItem[] = [
    {
      id: 1,
      src: '/img/home/achivie1.png',
      alt: 'Achievement 1',
      position: { top: '120px', left: '80px' }
    },
    {
      id: 2,
      src: '/img/home/achivie1.png',
      alt: 'Achievement 2',
      position: { top: '40px', left: '40%' }
    },
    {
      id: 3,
      src: '/img/home/achivie1.png',
      alt: 'Achievement 3',
      position: { top: '120px', right: '80px' }
    },
    {
      id: 4,
      src: '/img/home/achivie1.png',
      alt: 'Achievement 4',
      position: { bottom: '50px', left: '200px' }
    },
    {
      id: 5,
      src: '/img/home/achivie1.png',
      alt: 'Achievement 5',
      position: { bottom: '50px', right: '200px' }
    }
  ];

  useEffect(() => {
    const checkVisibility = () => {
      const guideSection = document.getElementById('guide');
      if (!guideSection) return;

      const rect = guideSection.getBoundingClientRect();
      const windowHeight = window.innerHeight;

      if (rect.top < windowHeight * 0.8) {
        guideSection.classList.add('animate');
      }
    };

    checkVisibility();

    window.addEventListener('scroll', checkVisibility);

    return () => {
      window.removeEventListener('scroll', checkVisibility);
    };
  }, []);

  const handleGuideClick = (guide: GuideKey) => {
    setActiveGuide(guide);
  };

  // ACHIVIE

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent, item: HTMLDivElement): void => {
      const rect = item.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      const xPercent = x / rect.width - 0.5;
      const yPercent = y / rect.height - 0.5;

      const tiltX = yPercent * 10;
      const tiltY = xPercent * -10;

      item.style.animation = 'none';
      item.style.transform = `perspective(1000px) rotateX(${tiltX}deg) rotateY(${tiltY}deg) scale(1.1)`;
    };

    const handleMouseLeave = (item: HTMLDivElement): void => {
      item.style.animation = '';
      item.style.transform = '';
    };

    const items = itemsRef.current;

    if (items && items.length > 0) {
      items.forEach((item) => {
        if (item) {
          const itemWithHandlers = item as DivWithHandlers;

          const moveHandler = (e: MouseEvent) => handleMouseMove(e, item);
          const leaveHandler = () => handleMouseLeave(item);

          item.addEventListener('mousemove', moveHandler);
          item.addEventListener('mouseleave', leaveHandler);

          // Gán handler vào DOM element (dạng mở rộng)
          itemWithHandlers._moveHandler = moveHandler;
          itemWithHandlers._leaveHandler = leaveHandler;
        }
      });
    }

    return () => {
      if (items && items.length > 0) {
        items.forEach((item) => {
          if (item) {
            const itemWithHandlers = item as DivWithHandlers;

            if (itemWithHandlers._moveHandler) item.removeEventListener('mousemove', itemWithHandlers._moveHandler);
            if (itemWithHandlers._leaveHandler) item.removeEventListener('mouseleave', itemWithHandlers._leaveHandler);
          }
        });
      }
    };
  }, []);

  return (
    <div className={cx('wrapper')}>
      {/* NAV */}
      <div className={cx('nav')}>
        {/* Logo */}
        <div className={cx('logo')}>
          <Link href='/'>
            <Image src='/img/logo_word.png' alt='logo' height={60} width={200} />
          </Link>
        </div>

        {/* Nav Items */}
        <div className={cx('nav-list')}>
          <div className={cx('nav-item')}>
            <Link href='#home'>Home</Link>
          </div>
          <div className={cx('nav-item')}>
            <Link href='/distributor/order'>Order</Link>
          </div>
          <div className={cx('nav-item')}>
            <Link href='/distributor/hatchery'>Hatchary</Link>
          </div>
          <div className={cx('nav-item')}>
            <Link href='#report'>Report</Link>
          </div>
        </div>

        {/* Search + User */}
        <div className={cx('nav-actions')}>
          <form>
            <svg width='20' height='20' viewBox='0 0 20 20'>
              <path
                fillRule='evenodd'
                clipRule='evenodd'
                d='M3.04175 9.37363C3.04175 5.87693 5.87711 3.04199 9.37508 3.04199C12.8731 3.04199 15.7084 5.87693 15.7084 9.37363C15.7084 12.8703 12.8731 15.7053 9.37508 15.7053C5.87711 15.7053 3.04175 12.8703 3.04175 9.37363ZM9.37508 1.54199C5.04902 1.54199 1.54175 5.04817 1.54175 9.37363C1.54175 13.6991 5.04902 17.2053 9.37508 17.2053C11.2674 17.2053 13.003 16.5344 14.357 15.4176L17.177 18.238C17.4699 18.5309 17.9448 18.5309 18.2377 18.238C18.5306 17.9451 18.5306 17.4703 18.2377 17.1774L15.418 14.3573C16.5365 13.0033 17.2084 11.2669 17.2084 9.37363C17.2084 5.04817 13.7011 1.54199 9.37508 1.54199Z'
                fill=''
              />
            </svg>
            <input type='text' placeholder='Search or type command...' />
            <button type='button'>
              <span>⌘</span>
              <span>K</span>
            </button>
          </form>

          <div className=''>
            <NotificationDropdown />
          </div>
          <div className=' text-white'>
            <UserDropdown />
          </div>
        </div>
      </div>

      {/* HERO */}

      <div className={`text-white items-center mx-auto px-6 py-10 space-y-10 ${cx('hero')}`} id='home'>
        {/* Summary Section */}
        <div>
          <h1 className='font-semibold text-[#7f8eff] border-b-2 border-[#7f8eff] w-fit pb-1 mb-4'>Summary</h1>
          <div className='space-y-2 text-lg'>
            <div>
              🌡️ Current Temperature: <span className='font-semibold'>28°C</span> (IDEAL)
            </div>
            <div>
              📦 Today's Order: <span className='font-semibold'>15</span> (10 white shrimp, 5 black tiger shrimp)
            </div>
            <div>
              💰 Revenue: <span className='font-semibold'>75,000,000 VND</span>
            </div>
            <div className='text-yellow-400'>⚠️ Warning: Tiger Shrimp Lot #123 expires in 2 days!</div>
          </div>
        </div>

        {/* Shipping Section */}
        <div>
          <h1 className='text-2xl font-semibold text-[#7f8eff] border-b-2 border-[#7f8eff] w-fit pb-1 mb-4'>
            🚛 Shipping
          </h1>
          <div className='space-y-2 text-lg'>
            <div>- Đơn #456 (Cần Thơ → Bạc Liêu | 3h/6h)</div>
            <div>- Đơn #789 (Cần Thơ → Sóc Trăng | 4h/8h)</div>
          </div>
        </div>
      </div>

      {/* FEATURES */}
      <section className={cx('features')} id='features'>
        <div className={cx('features-title')}>FEATURES</div>

        <div className={cx('horizontal-line')} />
        <div className={cx('vertical-line')} />

        <div
          className={cx(
            'features-list',
            'grid',
            'grid-cols-1',
            'gap-10',
            'w-full',
            'lg:grid-cols-3',
            'lg:w-5/6',
            'md:grid-cols-2',
            'sm:grid-cols-2'
          )}
        >
          <div className={cx('features-item', 'user-feature')}>
            <div className={cx('features-tag')}>ORDER</div>
            <div className={cx('features-name')}>Features 1</div>
            <div className={cx('features-description')}>
              Exercitationem omnis doloremque quasi. Architecto magni officia consequatur, fugiat totam iste
              perspiciatis. Tempora rem labore eaque in beatae, repellendus nulla modi.
            </div>

            <Link href={'/distributor/order'} className={cx('features-action')}>
              Explore
              <MoveRight size={30} />
            </Link>
          </div>

          <div className={cx('features-item', 'system-feature')}>
            <div className={cx('features-tag')}>HATCHARY</div>
            <div className={cx('features-name')}>Features 2</div>
            <div className={cx('features-description')}>
              Exercitationem omnis doloremque quasi. Architecto magni officia consequatur, fugiat totam iste
              perspiciatis. Tempora rem labore eaque in beatae, repellendus nulla modi.
            </div>

            <Link href={'/distributor/hatchary'} className={cx('features-action')}>
              Explore
              <MoveRight size={30} />
            </Link>
          </div>

          <div className={cx('features-item', 'user-feature', 'lg:col-span-1', 'md:col-span-2')}>
            <div className={cx('features-tag')}>REPORT</div>
            <div className={cx('features-name')}>Features 1</div>
            <div className={cx('features-description')}>
              Exercitationem omnis doloremque quasi. Architecto magni officia consequatur, fugiat totam iste
              perspiciatis. Tempora rem labore eaque in beatae, repellendus nulla modi.
            </div>

            <Link href={'#'} className={cx('features-action')}>
              Explore
              <MoveRight size={30} />
            </Link>
          </div>
        </div>

        {/* GUIDE - còn sửa */}
        <div className={cx('guide', 'animate')} id='guide'>
          <div className={cx('guide-left')}>
            <div
              className={cx('guide-item', { active: activeGuide === 'document' })}
              onClick={() => handleGuideClick('document')}
            >
              <div className={cx('guide-icon')}>
                <FileText size={24} />
              </div>
              <div className={cx('guide-name')}>DOCUMENT</div>
            </div>

            <div
              className={cx('guide-item', { active: activeGuide === 'guide' })}
              onClick={() => handleGuideClick('guide')}
            >
              <div className={cx('guide-icon')}>
                <BookOpen size={24} />
              </div>
              <div className={cx('guide-name')}>GUIDE</div>
            </div>

            <div
              className={cx('guide-item', { active: activeGuide === 'app' })}
              onClick={() => handleGuideClick('app')}
            >
              <div className={cx('guide-icon')}>
                <Smartphone size={24} />
              </div>
              <div className={cx('guide-name')}>APP</div>
            </div>
          </div>

          <div className={cx('guide-right')}>
            <div ref={guideRefs.document} className={cx('guide-content', { active: activeGuide === 'document' })}>
              <div className={cx('guide-title')}>A peer-reviewed study presenting novel findings.</div>
              <div className={cx('guide-image')}>
                <Image
                  src='/img/home/document-preview.png'
                  alt='Document Preview'
                  width={400}
                  height={300}
                  className={cx('guide-preview')}
                />
              </div>
            </div>

            <div ref={guideRefs.guide} className={cx('guide-content', { active: activeGuide === 'guide' })}>
              <div className={cx('guide-title')}>TracePost LarvaeChain User Guide</div>
              <div className={cx('guide-image')}>
                <Image
                  src='/img/home/guide-preview.png'
                  alt='Guide Preview'
                  width={400}
                  height={300}
                  className={cx('guide-preview')}
                />
              </div>
            </div>

            <div ref={guideRefs.app} className={cx('guide-content', { active: activeGuide === 'app' })}>
              <div className={cx('guide-title')}>Mobile application for tracking and monitoring.</div>
              <div className={cx('guide-image')}>
                <Image
                  src='/img/home/app-preview.png'
                  alt='App Preview'
                  width={400}
                  height={300}
                  className={cx('guide-preview')}
                />
              </div>
            </div>
          </div>
        </div>

        {/* --- */}
      </section>

      {/* ENTERPRISE */}
      <section className={cx('enterprise')} id='enterprise'>
        <div className={cx('horizontal-line')} />

        <div className={cx('enterprise-title')}>
          <div className={cx('vertical-line')} />
          Enterprise <span>Partners</span>
        </div>

        <div className={cx('enterprises-list')}>
          <div className={cx('enterprises-item', 'fpt')}>
            <Image src='/img/home/fpt.png' alt='fpt' width={60} height={10} />
            <div className={cx('enterprises-name')}>FPT UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'ctu')}>
            <Image src='/img/home/ctu.png' alt='ctu' width={45} height={45} />
            <div className={cx('enterprises-name')}>CAN THO UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'fpt')}>
            <Image src='/img/home/fpt.png' alt='fpt' width={60} height={10} />
            <div className={cx('enterprises-name')}>FPT UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'ctu')}>
            <Image src='/img/home/ctu.png' alt='ctu' width={45} height={45} />
            <div className={cx('enterprises-name')}>CAN THO UNIVERSITY</div>
          </div>
          <div className={cx('enterprises-item', 'fpt')}>
            <Image src='/img/home/fpt.png' alt='fpt' width={60} height={10} />
            <div className={cx('enterprises-name')}>FPT UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'ctu')}>
            <Image src='/img/home/ctu.png' alt='ctu' width={45} height={45} />
            <div className={cx('enterprises-name')}>CAN THO UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'fpt')}>
            <Image src='/img/home/fpt.png' alt='fpt' width={60} height={10} />
            <div className={cx('enterprises-name')}>FPT UNIVERSITY</div>
          </div>

          <div className={cx('enterprises-item', 'ctu')}>
            <Image src='/img/home/ctu.png' alt='ctu' width={45} height={45} />
            <div className={cx('enterprises-name')}>CAN THO UNIVERSITY</div>
          </div>
        </div>

        <div className={cx('horizontal-line', 'bottom-line')} />

        {/* ACHIVIE */}
        <div className={cx('achive')}>
          {/* Title in center */}
          <h4 className={cx('achive__title')}>OUR ACHIVIE</h4>

          {/* Images with positioning to match the layout */}
          {images.map((image, index) => (
            <div
              key={image.id}
              ref={(el) => {
                itemsRef.current[index] = el;
              }}
              className={cx('achive__item')}
              style={{
                ...image.position
              }}
            >
              <Image src={image.src} alt={image.alt} width={350} height={260} className={cx('achive__image')} />
            </div>
          ))}
        </div>
      </section>

      {/* FOOTER */}
      <footer>
        <Image src={'/img/home/footer_logo.png'} alt='logo' width={800} height={200} />

        <div className={cx('divider')} />

        <div className={cx('footer-container')}>
          <div className={cx('footer-info')}>
            <div className={cx('footer-description')}>2025 Fun is 9h. All Rights Reserved</div>
            <div className={cx('footer-social')}>
              <Link href={'#'}>
                <GithubIcon size={25} />
              </Link>

              <Link href={'#'}>
                <YoutubeIcon size={30} />
              </Link>

              <Link href={'mailto:'}>
                <Mail size={25} />
              </Link>
            </div>
          </div>

          <div className={cx('footer-content')}>
            <div className={cx('footer-navi')}>
              <div className={cx('navi-title')}>NAVIGATION</div>
              <div className={cx('navi-list')}>
                <Link href={'/distributor'} className={cx('navi-item')}>
                  HOME
                </Link>

                <Link href={'/distributor/order'} className={cx('navi-item')}>
                  ORDER
                </Link>

                <Link href={'/distributor/hatchary'} className={cx('navi-item')}>
                  HATCHARY
                </Link>
                <Link href={'#report'} className={cx('navi-item')}>
                  REPORT
                </Link>
              </div>
            </div>

            <div className={cx('footer-resources')}>
              <Link href={'#'} className={cx('resources-title')}>
                RESOURCES
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
export default Home;
