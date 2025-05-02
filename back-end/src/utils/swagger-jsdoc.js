/**
 * @swagger
 * /api/v1/auth/register:
 *   post:
 *     summary: Register a new user
 *     tags: [Auth]
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - name
 *               - email
 *               - password
 *             properties:
 *               name:
 *                 type: string
 *                 example: John Doe
 *               email:
 *                 type: string
 *                 format: email
 *                 example: john@example.com
 *               password:
 *                 type: string
 *                 format: password
 *                 example: password123
 *               role:
 *                 type: string
 *                 enum: [admin, manager, shipper, warehouse, customs, user]
 *                 default: user
 *                 example: manager
 *     responses:
 *       201:
 *         description: User successfully registered
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 token:
 *                   type: string
 *                   example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
 *                 data:
 *                   $ref: '#/components/schemas/User'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/auth/login:
 *   post:
 *     summary: Login to the system
 *     tags: [Auth]
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - email
 *               - password
 *             properties:
 *               email:
 *                 type: string
 *                 format: email
 *                 example: john@example.com
 *               password:
 *                 type: string
 *                 format: password
 *                 example: password123
 *     responses:
 *       200:
 *         description: User successfully logged in
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 token:
 *                   type: string
 *                   example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
 *                 data:
 *                   $ref: '#/components/schemas/User'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         description: Invalid credentials
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/Error'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/auth/me:
 *   get:
 *     summary: Get current user information
 *     tags: [Auth]
 *     security:
 *       - bearerAuth: []
 *     responses:
 *       200:
 *         description: Current user information
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   $ref: '#/components/schemas/User'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/shipments:
 *   get:
 *     summary: Get all shipments
 *     tags: [Shipments]
 *     responses:
 *       200:
 *         description: List of all shipments
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 count:
 *                   type: integer
 *                   example: 2
 *                 data:
 *                   type: array
 *                   items:
 *                     $ref: '#/components/schemas/Shipment'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *   post:
 *     summary: Create a new shipment
 *     tags: [Shipments]
 *     security:
 *       - bearerAuth: []
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - sender
 *               - receiver
 *               - origin
 *               - destination
 *               - product
 *             properties:
 *               sender:
 *                 type: string
 *                 example: Company A
 *               receiver:
 *                 type: string
 *                 example: Company B
 *               origin:
 *                 type: string
 *                 example: Shanghai, China
 *               destination:
 *                 type: string
 *                 example: Rotterdam, Netherlands
 *               product:
 *                 type: string
 *                 example: Electronics
 *               quantity:
 *                 type: number
 *                 example: 100
 *               unit:
 *                 type: string
 *                 example: kg
 *               description:
 *                 type: string
 *                 example: Consumer electronics shipment
 *     responses:
 *       201:
 *         description: Shipment successfully created
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   $ref: '#/components/schemas/Shipment'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/shipments/{id}:
 *   get:
 *     summary: Get a shipment by ID
 *     tags: [Shipments]
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: Shipment details
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   $ref: '#/components/schemas/Shipment'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *   put:
 *     summary: Update a shipment
 *     tags: [Shipments]
 *     security:
 *       - bearerAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               sender:
 *                 type: string
 *               receiver:
 *                 type: string
 *               origin:
 *                 type: string
 *               destination:
 *                 type: string
 *               product:
 *                 type: string
 *               quantity:
 *                 type: number
 *               unit:
 *                 type: string
 *               status:
 *                 type: string
 *                 enum: [CREATED, IN_TRANSIT, DELIVERED, CUSTOMS_CLEARED, DAMAGED, IN_WAREHOUSE, RETURNED]
 *               description:
 *                 type: string
 *     responses:
 *       200:
 *         description: Shipment successfully updated
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   $ref: '#/components/schemas/Shipment'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *   delete:
 *     summary: Delete a shipment
 *     tags: [Shipments]
 *     security:
 *       - bearerAuth: []
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: Shipment successfully deleted
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   type: object
 *                   example: {}
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/shipments/{id}/events:
 *   get:
 *     summary: Get events for a shipment
 *     tags: [Shipments, Events]
 *     parameters:
 *       - in: path
 *         name: id
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: List of events for the shipment
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 count:
 *                   type: integer
 *                   example: 3
 *                 data:
 *                   type: array
 *                   items:
 *                     $ref: '#/components/schemas/Event'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/events/{shipmentId}:
 *   get:
 *     summary: Get events for a shipment
 *     tags: [Events]
 *     parameters:
 *       - in: path
 *         name: shipmentId
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: List of events for the shipment
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 count:
 *                   type: integer
 *                   example: 3
 *                 data:
 *                   type: array
 *                   items:
 *                     $ref: '#/components/schemas/Event'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *   post:
 *     summary: Record a new event for a shipment
 *     tags: [Events]
 *     security:
 *       - bearerAuth: []
 *     parameters:
 *       - in: path
 *         name: shipmentId
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             required:
 *               - type
 *               - location
 *             properties:
 *               type:
 *                 type: string
 *                 enum: [PICKUP, DELIVERY, CUSTOMS_CLEARANCE, WAREHOUSE_IN, WAREHOUSE_OUT, DAMAGED, RETURNED]
 *                 example: PICKUP
 *               location:
 *                 type: string
 *                 example: Shanghai Port
 *               data:
 *                 type: object
 *                 example:
 *                   temperature: 25C
 *                   humidity: 60%
 *                   handler: John Smith
 *               timestamp:
 *                 type: string
 *                 format: date-time
 *                 example: 2023-01-01T00:00:00.000Z
 *     responses:
 *       201:
 *         description: Event successfully recorded
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   $ref: '#/components/schemas/Event'
 *       400:
 *         $ref: '#/components/responses/ValidationError'
 *       401:
 *         $ref: '#/components/responses/UnauthorizedError'
 *       403:
 *         $ref: '#/components/responses/ForbiddenError'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/tracing/verify/{shipmentId}:
 *   get:
 *     summary: Verify a shipment on the blockchain
 *     tags: [Tracing]
 *     parameters:
 *       - in: path
 *         name: shipmentId
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: Shipment verification result
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 data:
 *                   type: object
 *                   properties:
 *                     shipment:
 *                       $ref: '#/components/schemas/Shipment'
 *                     blockchain:
 *                       $ref: '#/components/schemas/BlockchainVerification'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 *
 * /api/v1/tracing/events/{shipmentId}:
 *   get:
 *     summary: Get blockchain events for a shipment
 *     tags: [Tracing]
 *     parameters:
 *       - in: path
 *         name: shipmentId
 *         required: true
 *         schema:
 *           type: string
 *         description: Shipment ID
 *     responses:
 *       200:
 *         description: List of blockchain events for the shipment
 *         content:
 *           application/json:
 *             schema:
 *               type: object
 *               properties:
 *                 success:
 *                   type: boolean
 *                   example: true
 *                 count:
 *                   type: integer
 *                   example: 3
 *                 data:
 *                   type: array
 *                   items:
 *                     $ref: '#/components/schemas/BlockchainEvent'
 *       400:
 *         description: Blockchain verification failed
 *         content:
 *           application/json:
 *             schema:
 *               $ref: '#/components/schemas/Error'
 *       404:
 *         $ref: '#/components/responses/NotFoundError'
 *       500:
 *         $ref: '#/components/responses/ServerError'
 */
