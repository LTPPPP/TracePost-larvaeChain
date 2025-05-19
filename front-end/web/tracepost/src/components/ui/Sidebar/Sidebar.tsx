'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { LogOut, LucideIcon } from 'lucide-react';
import { useRouter } from 'next/navigation';

import classNames from 'classnames/bind';
import styles from './Sidebar.module.scss';

const cx = classNames.bind(styles);

export interface MenuItem {
  icon: LucideIcon;
  name: string;
  link: string;
}

interface SidebarProps {
  menuItems: MenuItem[];
}

const Sidebar: React.FC<SidebarProps> = ({ menuItems }) => {
  const [isLocked, setIsLocked] = useState<boolean | null>(null);
  const [isVisible, setIsVisible] = useState<boolean>(true);
  let timeoutId: NodeJS.Timeout | null = null;

  const router = useRouter();

  useEffect(() => {
    const sidebarLock = localStorage.getItem('sidebarLock');
    if (sidebarLock !== null) {
      setIsLocked(sidebarLock === 'true');
    } else {
      setIsLocked(true);
    }
  }, []);

  useEffect(() => {
    if (isLocked !== null) {
      localStorage.setItem('sidebarLock', isLocked.toString());
    }
  }, [isLocked]);

  const hideSidebar = (): void => {
    if (!isLocked) {
      setIsVisible(false);
    }
  };

  const resetTimeout = (): void => {
    if (timeoutId) clearTimeout(timeoutId);
    if (!isLocked) {
      timeoutId = setTimeout(hideSidebar, 2000);
    }
  };

  useEffect(() => {
    resetTimeout();
    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isVisible, isLocked]);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent): void => {
      if (e.clientX < 50) {
        setIsVisible(true);
      }
    };
    window.addEventListener('mousemove', handleMouseMove);
    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, []);

  const handleLogout = () => {
    // handle logout
    router.replace('/login');
  };

  return (
    <div
      className={cx('wrapper', { hide: !isVisible })}
      onMouseEnter={() => setIsVisible(true)}
      onMouseMove={resetTimeout}
    >
      <div className={cx('content')}>
        <Image src='/img/logo_word.png' alt='logo' height={200} width={200} />
        <div className={cx('title')}>Menu</div>

        <div className={cx('list', 'flex-1')}>
          {menuItems.map((item, index) => {
            const IconComponent = item.icon;
            return (
              <div key={index} className={cx('item')}>
                <Link href={item.link}>
                  <IconComponent />
                  <span>{item.name}</span>
                </Link>
              </div>
            );
          })}
        </div>

        <div className={cx('title')}>Setting</div>
        <div className={cx('setting-bg')}>
          <div className={cx('item')}>
            Lock
            <label className={cx('toggle-switch')}>
              <input type='checkbox' checked={isLocked || false} onChange={() => setIsLocked((prev) => !prev)} />
              <span className={cx('slider')}></span>
            </label>
          </div>
        </div>
      </div>

      <div className={cx('action')}>
        <button className={cx('logout-btn')} onClick={handleLogout}>
          <LogOut />
          <span>LOGOUT</span>
        </button>
      </div>
    </div>
  );
};

export default Sidebar;
