'use client';

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';

import Typed from 'typed.js';
import { FileText, BookOpen, Smartphone, MoveRight, YoutubeIcon, GithubIcon, Mail } from 'lucide-react';

import styles from './home.module.scss';
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
  // HERO
  const typedTextRef = useRef<HTMLSpanElement | null>(null);

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

  // GUIDE
  useEffect(() => {
    const typed = new Typed(typedTextRef.current, {
      strings: [
        '<span style="color: var(--purple-color)">Trans</span><span style="color: var(--orange-color)">parency</span>',
        '<span style="color: var(--purple-color)">Shri</span><span style="color: var(--orange-color)">mp</span>',
        '<span style="color: var(--purple-color)">Trace</span><span style="color: var(--orange-color)">ability</span>'
      ],
      typeSpeed: 80,
      backSpeed: 50,
      backDelay: 1500,
      startDelay: 500,
      loop: true,
      cursorChar: '|',
      smartBackspace: true,
      showCursor: false,
      autoInsertCss: true,
      contentType: 'html'
    });

    return () => {
      typed.destroy();
    };
  }, []);

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
      <div className={cx('vertical-line')} />

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
            <span ref={typedTextRef}></span>
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
            <div className={cx('features-tag')}>USER</div>
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

          <div className={cx('features-item', 'system-feature')}>
            <div className={cx('features-tag')}>SHRIMPER</div>
            <div className={cx('features-name')}>Features 2</div>
            <div className={cx('features-description')}>
              Exercitationem omnis doloremque quasi. Architecto magni officia consequatur, fugiat totam iste
              perspiciatis. Tempora rem labore eaque in beatae, repellendus nulla modi.
            </div>

            <Link href={'#'} className={cx('features-action')}>
              Explore
              <MoveRight size={30} />
            </Link>
          </div>

          <div className={cx('features-item', 'user-feature', 'lg:col-span-1', 'md:col-span-2')}>
            <div className={cx('features-tag')}>USER</div>
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
          <h2 className={cx('achive__title')}>OUR ACHIVIE</h2>

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
        <Image src={'/img/home/footer_logo.png'} alt='logo' width={1048} height={220} />

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
                <Link href={'#home'} className={cx('navi-item')}>
                  HOME
                </Link>

                <Link href={'#features'} className={cx('navi-item')}>
                  FEATURES
                </Link>

                <Link href={'#enterprise'} className={cx('navi-item')}>
                  ENTERPRISE
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
