import ComponentCard from '@/components/common/ComponentCard';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import BasicTableOne from '@/components/tables/BasicTableOne';
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
        <ComponentCard title='Basic Table 1'>
          <BasicTableOne />
        </ComponentCard>
      </div>
    </div>
  );
}
