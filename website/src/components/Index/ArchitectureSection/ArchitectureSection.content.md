The Yunikorn History Server (YHS) is a standalone service that enhances the capabilities of the
Yunikorn Scheduler by providing long-term persistence of cluster operational data.
It achieves this by listening for events from the Yunikorn Scheduler and persisting them to a database.

![YHS Architecture](https://github.com/G-Research/yunikorn-history-server/raw/main/yhs-architecture.png)

YHS is composed of three main components:

1. **Event Collector:** This component is responsible for listening to events stream API from the Yunikorn Scheduler  
   and persisting them to the database.
   It ensures that all significant operations performed by the scheduler
   are recorded for future analysis.

2. **REST API:** This component serves as the interface for retrieving historical data.
   It provides endpoints that return data about past applications, resource usage, and other operational metrics. Also provides
   querying capabilities to filter and retrieve specific data.

3. **Web Frontend:** This component enhances the existing Yunikorn Web interface by providing additional features that utilize
   the historical data stored by YHS. It is loaded on the application page of Yunikorn Web.
   More details on the web component is available [here](https://github.com/G-Research/yunikorn-history-server/tree/main/web).

By integrating these components, YHS provides a comprehensive view of the historical operations of a Yunikorn-managed cluster,
enabling detailed analysis and insights.
