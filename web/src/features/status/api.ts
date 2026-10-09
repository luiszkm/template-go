import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";

export function useApiStatus() {
  return useQuery({
    queryKey: ["readyz"],
    queryFn: async () => {
      const { data, response } = await api.GET("/readyz");
      if (response.status !== 200 || !data) {
        throw new Error(`readyz answered ${response.status}`);
      }
      return data;
    },
    retry: false,
  });
}
