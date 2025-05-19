'use client';

import Sidebar, { MenuItem } from '@/components/ui/Sidebar/Sidebar';
import { LayoutDashboard, UserRound, ShoppingBasket, FolderPlus } from 'lucide-react';

import classNames from 'classnames/bind';
import styles from './Workspace.module.scss';
import Clock from '@/components/ui/Clock/Clock';

const cx = classNames.bind(styles);
function Workspace() {
  // MENU
  const menuItems: MenuItem[] = [
    {
      icon: LayoutDashboard,
      name: 'Workspace',
      link: '/workspace'
    },
    {
      icon: FolderPlus,
      name: 'Create Batch',
      link: '/create-batch'
    },
    {
      icon: ShoppingBasket,
      name: 'Order History',
      link: '/order-history'
    },
    {
      icon: UserRound,
      name: 'Profile',
      link: '/profile'
    }
  ];

  return (
    <div className={cx('wrapper')}>
      <Clock />
      <Sidebar menuItems={menuItems} />
      <div className={cx('container')}></div>
    </div>
  );
}

export default Workspace;
