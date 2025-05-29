import ComponentCard from '@/components/common/ComponentCard';
import PageBreadcrumb from '@/components/common/PageBreadCrumb';
import TableDistributor from '@/components/tables/TableDistributor';
import { Metadata } from 'next';
import React from 'react';

export const metadata: Metadata = {
  title: 'TracePost',
  description: 'A high-performance backend system for shrimp larvae traceability using blockchain technology.'
};

export default function DistributorTables() {
  return (
    <div>
      <PageBreadcrumb pageTitle='Distributor List' />
      <div className='space-y-6'>
        <ComponentCard title='Table Distributor'>
          <TableDistributor />
        </ComponentCard>
      </div>
    </div>
  );
}
