'use client';
import React from 'react';
import { useState } from 'react';
import { Table, TableBody, TableCell, TableHeader, TableRow } from '../ui/table';
import { EyeIcon } from '../../icons/index';
import Badge from '../ui/badge/Badge';
import Image from 'next/image';
import ToggleButton from '../ui/toggle/index';

interface Order {
  id: number;
  user: {
    image: string;
    name: string;
    role: string;
  };
  details: React.ReactNode;
  company: {
    images: string[];
  };
  email: string;
  status: boolean;
}

// Define the table data using the interface
const tableData: Order[] = [
  {
    id: 1,
    user: {
      image: '/images/user/user-17.jpg',
      name: 'Lindsey Curtis',
      role: ''
    },
    details: <EyeIcon />,
    company: {
      images: ['/images/user/user-22.jpg', '/images/user/user-23.jpg', '/images/user/user-24.jpg']
    },
    email: 'nguyengiachan.gr2020@gmail.com',
    status: true
  },
  {
    id: 2,
    user: {
      image: '/images/user/user-18.jpg',
      name: 'Kaiya George',
      role: ''
    },
    details: <EyeIcon />,
    company: {
      images: ['/images/user/user-25.jpg', '/images/user/user-26.jpg']
    },
    email: '24.9K',
    status: false
  },
  {
    id: 3,
    user: {
      image: '/images/user/user-17.jpg',
      name: 'Zain Geidt',
      role: ''
    },
    details: <EyeIcon />,
    company: {
      images: ['/images/user/user-27.jpg']
    },
    email: '12.7K',
    status: true
  },
  {
    id: 4,
    user: {
      image: '/images/user/user-20.jpg',
      name: 'Abram Schleifer',
      role: ''
    },
    details: <EyeIcon />,
    company: {
      images: ['/images/user/user-28.jpg', '/images/user/user-29.jpg', '/images/user/user-30.jpg']
    },
    email: '2.8K',
    status: true
  },
  {
    id: 5,
    user: {
      image: '/images/user/user-21.jpg',
      name: 'Carla George',
      role: ''
    },
    details: <EyeIcon />,
    company: {
      images: ['/images/user/user-31.jpg', '/images/user/user-32.jpg', '/images/user/user-33.jpg']
    },
    email: '4.5K',
    status: true
  }
];

export default function AcceptTable() {
  const [orders, setOrders] = useState<Order[]>(tableData);

  const handleToggle = (id: number, checked: boolean) => {
    const updated = orders.map((order) => (order.id === id ? { ...order, status: checked } : order));
    setOrders(updated);
  };

  return (
    <div className='overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/[0.05] dark:bg-white/[0.03]'>
      <div className='max-w-full overflow-x-auto'>
        <div className='min-w-[1102px]'>
          <Table>
            {/* Table Header */}
            <TableHeader className='border-b border-gray-100 dark:border-white/[0.05]'>
              <TableRow>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  User
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Detail
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Company
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400'
                >
                  Email
                </TableCell>
                <TableCell
                  isHeader
                  className='px-5 py-3 text-center font-medium text-gray-500 text-theme-xs dark:text-gray-400'
                >
                  Status
                </TableCell>
              </TableRow>
            </TableHeader>

            {/* Table Body */}
            <TableBody className='divide-y divide-gray-100 dark:divide-white/[0.05]'>
              {orders.map((order) => (
                <TableRow key={order.id}>
                  <TableCell className='px-5 py-4 sm:px-6 text-start'>
                    <div className='flex items-center gap-3'>
                      <div className='w-10 h-10 overflow-hidden rounded-full'>
                        <Image width={40} height={40} src={order.user.image} alt={order.user.name} />
                      </div>
                      <div>
                        <span className='block font-medium text-gray-800 text-theme-sm dark:text-white/90'>
                          {order.user.name}
                        </span>
                        <span className='block text-gray-500 text-theme-xs dark:text-gray-400'>{order.user.role}</span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className='px-4 py-3 text-gray-500 item-center mx-auto text-theme-sm dark:text-gray-400'>
                    {order.details}
                  </TableCell>
                  <TableCell className='px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400'>
                    <div className='flex -space-x-2'>
                      {order.company.images.map((teamImage, index) => (
                        <div
                          key={index}
                          className='w-6 h-6 overflow-hidden border-2 border-white rounded-full dark:border-gray-900'
                        >
                          <Image
                            width={24}
                            height={24}
                            src={teamImage}
                            alt={`Team member ${index + 1}`}
                            className='w-full'
                          />
                        </div>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className='px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400'>
                    <Badge size='sm'>{order.email}</Badge>
                  </TableCell>
                  <TableCell className='px-4 py-3 text-gray-500 text-theme-sm dark:text-gray-400'>
                    <div className='flex items-center justify-center gap-2'>
                      <button
                        type='button'
                        className='focus:outline-none text-white bg-green-600 hover:bg-green-800 focus:ring-4 focus:ring-green-300 font-medium rounded-lg text-sm px-5 py-2 me-2 mb-2 dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800'
                      >
                        Accept
                      </button>

                      <div className='w-px h-7 bg-gray-400'></div>

                      <button
                        type='button'
                        className='focus:outline-none text-white bg-red-600 hover:bg-red-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-5 py-2 me-2 mb-2 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-900'
                      >
                        Reject
                      </button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
