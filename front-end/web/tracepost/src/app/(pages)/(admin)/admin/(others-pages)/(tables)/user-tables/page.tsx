import ComponentCard from '@/components/common/ComponentCard';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import TableUser from '@/components/tables/TableUser';
import { Metadata } from 'next';
import React from 'react';

export const metadata: Metadata = {
  title: 'TracePost',
  description: 'A high-performance backend system for shrimp larvae traceability using blockchain technology.'
};

export default function BasicTables() {
  return (
    <div>
      <PageBreadcrumb pageTitle='User List' />
      <div className='space-y-6'>
        <ComponentCard title='Table User'>
          <TableUser />
        </ComponentCard>
      </div>
    </div>
  );
}
