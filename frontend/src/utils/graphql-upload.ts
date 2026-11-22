import { addAuthHeader } from "@/utils/auth-headers";
import { withBasePath } from "@/utils/base-path";

export type GraphQLUploadRequest = {
  query: string;
  variables: Record<string, any>;
  file: File;
  filePath: string;
  operationName?: string;
};

export async function executeMultipartGraphQL<T>({
  query,
  variables,
  file,
  filePath,
  operationName,
}: GraphQLUploadRequest): Promise<T> {
  const formData = new FormData();
  formData.append(
    "operations",
    JSON.stringify({
      query,
      variables,
      operationName,
    }),
  );
  formData.append("map", JSON.stringify({ "0": [filePath] }));
  formData.append("0", file);

  const response = await fetch(withBasePath("/api/query"), {
    method: "POST",
    credentials: "include",
    headers: addAuthHeader({}),
    body: formData,
  });

  const result = await response.json();
  if (!response.ok || result.errors?.length) {
    const message = result.errors?.[0]?.message ?? "";
    throw new Error(message);
  }

  return result.data as T;
}