import { defineSitemapEventHandler } from '#imports'

export default defineSitemapEventHandler(async () => {
  const backendUrl = '/api';
  const resp = await $fetch(`${backendUrl}/report/data/list`, {
    method: "GET",
    query: { page: "1", page_size: "65535" },
  });
  if (resp && resp.code === 0) {
    const reports = resp.data || [];
    return reports.map(report => ({
      loc: `/report/${report.id}`,
    }));
  }
  return []
})