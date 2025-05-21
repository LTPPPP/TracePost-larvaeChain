import ComponentCard from '@/components/common/ComponentCard';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import AcceptTable from '@/components/tables/AcceptTable';
import { Metadata } from 'next';
import React from 'react';

export const metadata: Metadata = {
  title: 'TracePost',
  description: 'A high-performance backend system for shrimp larvae traceability using blockchain technology.'
};

export default function AcceptTables() {
  return (
    <div>
      <PageBreadcrumb pageTitle='Accept List' />
      <div className='space-y-6'>
        <ComponentCard title='Table Users Want To Upgrade To Hatchary'>
          <AcceptTable />
        </ComponentCard>
      </div>
    </div>
  );
}
