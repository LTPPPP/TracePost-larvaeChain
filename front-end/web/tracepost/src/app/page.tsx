'use client';

import { useEffect, useRef, useState } from 'react';
import Image from 'next/image';
import Link from 'next/link';

import Typed from 'typed.js';
import { FileText, BookOpen, Smartphone, MoveRight } from 'lucide-react';

import styles from './Home.module.scss';
import classNames from 'classnames/bind';
const cx = classNames.bind(styles);

function Home() {
  const [activeGuide, setActiveGuide] = useState('document');
  const guideRefs = {
    document: useRef(null),
    guide: useRef(null),
    app: useRef(null)
  };
  const typedTextRef = useRef(null);

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

    // Check ngay khi component mount
    checkVisibility();

    // Add listener cho scroll event
    window.addEventListener('scroll', checkVisibility);

    // Cleanup function
    return () => {
      window.removeEventListener('scroll', checkVisibility);
    };
  }, []);

  const handleGuideClick = (guide: any) => {
    setActiveGuide(guide);

    // scroll functionality
    // if (guideRefs[guide].current) {
    //   guideRefs[guide].current.scrollIntoView({
    //     behavior: 'smooth',
    //     block: 'start'
    //   });
    // }
  };

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
    </div>
  );
}
export default Home;
