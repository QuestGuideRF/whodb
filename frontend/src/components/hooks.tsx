import type { SourceObjectRefInput } from "@graphql";
import {useCallback} from "react";
import * as desktopService from "../services/desktop";
import {isDesktopApp} from "../utils/external-links";
import {addAuthHeader} from "../utils/auth-headers";
import {withBasePath} from "../utils/base-path";

export const useExportToCSV = (
    objectRef: SourceObjectRefInput | undefined,
    fileBaseName: string,
    selectedOnly: boolean = false,
    delimiter: string = ',',
    selectedRows?: Record<string, any>[],
    format: 'csv' | 'excel' | 'ndjson' = 'csv'
) => {
    return useCallback(async () => {
      try {

        const requestBody: any = {
          ref: objectRef,
          fileBaseName,
          delimiter,
          format,
        };

        if (selectedOnly && selectedRows && selectedRows.length > 0) {

          if (selectedRows.length > 1000) {
            console.warn(`Exporting ${selectedRows.length} rows. Large selections may be slow.`);
          }
          requestBody.selectedRows = selectedRows;
        }

        const response = await fetch(withBasePath('/api/export'), {
          method: 'POST',
          credentials: 'include',
          headers: addAuthHeader({
            'Accept': format === 'excel' ? 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' : format === 'ndjson' ? 'application/x-ndjson' : 'text/csv',
            'Content-Type': 'application/json',
          }),
          body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
          throw new Error(await response.text());
        }

        const contentDisposition = response.headers.get('Content-Disposition');
        const extension = format === 'excel' ? 'xlsx' : format === 'ndjson' ? 'ndjson' : 'csv';
        let filename = `${fileBaseName}.${extension}`;
        if (contentDisposition) {
          const filenameMatch = contentDisposition.match(/filename="(.+)"/);
          if (filenameMatch) {
            filename = filenameMatch[1];
          }
        }

        const blob = await response.blob();

        if (isDesktopApp()) {
          const arrayBuffer = await blob.arrayBuffer();
          const data = new Uint8Array(arrayBuffer);
          const savedPath = await desktopService.saveBinaryFile(data, filename);
          if (!savedPath) {
            console.error('Export: Save dialog was cancelled or failed');

            return;
          }
        } else {

          const downloadUrl = window.URL.createObjectURL(blob);
          const link = document.createElement('a');
          link.href = downloadUrl;
          link.download = filename;
          link.style.display = 'none';
          document.body.appendChild(link);

          link.click();

          setTimeout(() => {
            document.body.removeChild(link);
            window.URL.revokeObjectURL(downloadUrl);
          }, 100);
        }
      } catch (error) {
        throw error;
      }
    }, [objectRef, fileBaseName, selectedOnly, delimiter, selectedRows, format]);
};