'use client';
import { useState } from 'react';
import Link from 'next/link';
import { Button } from '@/app/ui/button';
import { createStrategy } from '@/app/lib/actions';

export default function Form() {
  const [fileName, setFileName] = useState(null)
  const handleFileChange = (e) => {
    const file = e.target.files?.[0];
    if (file) {
      setFileName(file.name);
    } else {
      setFileName(null);
    }
  };

  return (
    <form action={createStrategy}>
      <div className="rounded-md bg-gray-50 p-4 md:p-6">
        {/* Strategy Name */}
        <div className="mb-4">
          <label htmlFor="name" className="mb-2 block text-sm font-medium">
            Name a strategy
          </label>
          <div className="relative mt-2 rounded-md">
            <div className="relative">
              <input
                id="name"
                name="name"
                type="text"
                placeholder="Enter the name for a strategy"
                className="peer block w-full rounded-md border border-gray-20 pl-10 text-sm outline-2 placeholder:text-gray-500"
                required
              />
            </div>
          </div>
        </div>

        {/* File Upload */}
        <fieldset>
          <legend className="mb-2 block text-sm font-medium">
            Upload CSV file
          </legend>
          <label
            htmlFor="file"
            className="block rounded-md border border-gray-200 bg-white px-[14px] py-3 text-center cursor-pointer"
          >
            {fileName || 'Upload CSV'}
          </label>
          <input
            id="file"
            name="file"
            type="file"
            accept=".csv"
            className="hidden"
            onChange={handleFileChange}
          />
        </fieldset>
      </div>
      <div className="mt-6 flex justify-end gap-4">
        <Link
          href="/dashboard/invoices"
          className="flex h-10 items-center rounded-lg bg-gray-100 px-4 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-200"
        >
          Cancel
        </Link>
        <Button type="submit">Create Strategy</Button>
      </div>
    </form>
  );
}
