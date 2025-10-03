import Foundation
import HealthKit
import WatchKit

class WatchHealthManager: NSObject, ObservableObject {
    static let shared = WatchHealthManager()
    
    private let healthStore = HKHealthStore()
    
    @Published var healthMetrics = HealthMetrics()
    @Published var isAuthorized = false
    @Published var isTrackingWorkout = false
    @Published var currentWorkout: HKWorkout?
    
    private var workoutSession: HKWorkoutSession?
    private var workoutBuilder: HKLiveWorkoutBuilder?
    
    // Health data types we want to read
    private let readTypes: Set<HKObjectType> = [
        HKObjectType.quantityType(forIdentifier: .stepCount)!,
        HKObjectType.quantityType(forIdentifier: .activeEnergyBurned)!,
        HKObjectType.quantityType(forIdentifier: .distanceWalkingRunning)!,
        HKObjectType.quantityType(forIdentifier: .heartRate)!,
        HKObjectType.quantityType(forIdentifier: .appleExerciseTime)!,
        HKObjectType.workoutType()
    ]
    
    // Health data types we want to write
    private let writeTypes: Set<HKSampleType> = [
        HKObjectType.quantityType(forIdentifier: .stepCount)!,
        HKObjectType.quantityType(forIdentifier: .activeEnergyBurned)!,
        HKObjectType.quantityType(forIdentifier: .distanceWalkingRunning)!,
        HKObjectType.workoutType()
    ]
    
    override init() {
        super.init()
        requestHealthKitAuthorization()
    }
    
    // MARK: - Authorization
    
    private func requestHealthKitAuthorization() {
        guard HKHealthStore.isHealthDataAvailable() else {
            print("HealthKit is not available on this device")
            return
        }
        
        healthStore.requestAuthorization(toShare: writeTypes, read: readTypes) { [weak self] success, error in
            DispatchQueue.main.async {
                self?.isAuthorized = success
                if success {
                    self?.setupHealthDataObservers()
                    self?.loadTodayHealthData()
                }
                
                if let error = error {
                    print("HealthKit authorization error: \(error)")
                }
            }
        }
    }
    
    // MARK: - Health Data Observers
    
    private func setupHealthDataObservers() {
        setupStepsObserver()
        setupCaloriesObserver()
        setupHeartRateObserver()
        setupDistanceObserver()
    }
    
    private func setupStepsObserver() {
        guard let stepType = HKQuantityType.quantityType(forIdentifier: .stepCount) else { return }
        
        let query = HKObserverQuery(sampleType: stepType, predicate: nil) { [weak self] _, _, error in
            if let error = error {
                print("Steps observer error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                self?.loadTodaySteps()
            }
        }
        
        healthStore.execute(query)
        healthStore.enableBackgroundDelivery(for: stepType, frequency: .immediate) { success, error in
            if let error = error {
                print("Background delivery setup error: \(error)")
            }
        }
    }
    
    private func setupCaloriesObserver() {
        guard let calorieType = HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned) else { return }
        
        let query = HKObserverQuery(sampleType: calorieType, predicate: nil) { [weak self] _, _, error in
            if let error = error {
                print("Calories observer error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                self?.loadTodayCalories()
            }
        }
        
        healthStore.execute(query)
        healthStore.enableBackgroundDelivery(for: calorieType, frequency: .immediate) { success, error in
            if let error = error {
                print("Background delivery setup error: \(error)")
            }
        }
    }
    
    private func setupHeartRateObserver() {
        guard let heartRateType = HKQuantityType.quantityType(forIdentifier: .heartRate) else { return }
        
        let query = HKObserverQuery(sampleType: heartRateType, predicate: nil) { [weak self] _, _, error in
            if let error = error {
                print("Heart rate observer error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                self?.loadLatestHeartRate()
            }
        }
        
        healthStore.execute(query)
    }
    
    private func setupDistanceObserver() {
        guard let distanceType = HKQuantityType.quantityType(forIdentifier: .distanceWalkingRunning) else { return }
        
        let query = HKObserverQuery(sampleType: distanceType, predicate: nil) { [weak self] _, _, error in
            if let error = error {
                print("Distance observer error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                self?.loadTodayDistance()
            }
        }
        
        healthStore.execute(query)
    }
    
    // MARK: - Health Data Loading
    
    func loadTodayHealthData() {
        loadTodaySteps()
        loadTodayCalories()
        loadTodayDistance()
        loadLatestHeartRate()
        loadTodayExerciseTime()
    }
    
    private func loadTodaySteps() {
        guard let stepType = HKQuantityType.quantityType(forIdentifier: .stepCount) else { return }
        
        let predicate = HKQuery.predicateForSamples(withStart: Calendar.current.startOfDay(for: Date()),
                                                   end: Date(),
                                                   options: .strictStartDate)
        
        let query = HKStatisticsQuery(quantityType: stepType,
                                     quantitySamplePredicate: predicate,
                                     options: .cumulativeSum) { [weak self] _, result, error in
            if let error = error {
                print("Steps query error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                let steps = result?.sumQuantity()?.doubleValue(for: .count()) ?? 0
                self?.healthMetrics.steps = Int(steps)
                self?.checkStepMilestones(Int(steps))
            }
        }
        
        healthStore.execute(query)
    }
    
    private func loadTodayCalories() {
        guard let calorieType = HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned) else { return }
        
        let predicate = HKQuery.predicateForSamples(withStart: Calendar.current.startOfDay(for: Date()),
                                                   end: Date(),
                                                   options: .strictStartDate)
        
        let query = HKStatisticsQuery(quantityType: calorieType,
                                     quantitySamplePredicate: predicate,
                                     options: .cumulativeSum) { [weak self] _, result, error in
            if let error = error {
                print("Calories query error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                let calories = result?.sumQuantity()?.doubleValue(for: .kilocalorie()) ?? 0
                self?.healthMetrics.calories = Int(calories)
                self?.checkCalorieMilestones(Int(calories))
            }
        }
        
        healthStore.execute(query)
    }
    
    private func loadTodayDistance() {
        guard let distanceType = HKQuantityType.quantityType(forIdentifier: .distanceWalkingRunning) else { return }
        
        let predicate = HKQuery.predicateForSamples(withStart: Calendar.current.startOfDay(for: Date()),
                                                   end: Date(),
                                                   options: .strictStartDate)
        
        let query = HKStatisticsQuery(quantityType: distanceType,
                                     quantitySamplePredicate: predicate,
                                     options: .cumulativeSum) { [weak self] _, result, error in
            if let error = error {
                print("Distance query error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                let distance = result?.sumQuantity()?.doubleValue(for: .meter()) ?? 0
                self?.healthMetrics.distance = distance
            }
        }
        
        healthStore.execute(query)
    }
    
    private func loadLatestHeartRate() {
        guard let heartRateType = HKQuantityType.quantityType(forIdentifier: .heartRate) else { return }
        
        let sortDescriptor = NSSortDescriptor(key: HKSampleSortIdentifierEndDate, ascending: false)
        let query = HKSampleQuery(sampleType: heartRateType,
                                 predicate: nil,
                                 limit: 1,
                                 sortDescriptors: [sortDescriptor]) { [weak self] _, samples, error in
            if let error = error {
                print("Heart rate query error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                if let sample = samples?.first as? HKQuantitySample {
                    let heartRate = sample.quantity.doubleValue(for: HKUnit(from: "count/min"))
                    self?.healthMetrics.heartRate = Int(heartRate)
                    self?.checkHeartRateAlerts(Int(heartRate))
                }
            }
        }
        
        healthStore.execute(query)
    }
    
    private func loadTodayExerciseTime() {
        guard let exerciseType = HKQuantityType.quantityType(forIdentifier: .appleExerciseTime) else { return }
        
        let predicate = HKQuery.predicateForSamples(withStart: Calendar.current.startOfDay(for: Date()),
                                                   end: Date(),
                                                   options: .strictStartDate)
        
        let query = HKStatisticsQuery(quantityType: exerciseType,
                                     quantitySamplePredicate: predicate,
                                     options: .cumulativeSum) { [weak self] _, result, error in
            if let error = error {
                print("Exercise time query error: \(error)")
                return
            }
            
            DispatchQueue.main.async {
                let exerciseTime = result?.sumQuantity()?.doubleValue(for: .minute()) ?? 0
                self?.healthMetrics.activeMinutes = Int(exerciseTime)
            }
        }
        
        healthStore.execute(query)
    }
    
    // MARK: - Workout Management
    
    func startWorkout(type: HKWorkoutActivityType = .walking) {
        guard !isTrackingWorkout else { return }
        
        let configuration = HKWorkoutConfiguration()
        configuration.activityType = type
        configuration.locationType = .outdoor
        
        do {
            workoutSession = try HKWorkoutSession(healthStore: healthStore, configuration: configuration)
            workoutBuilder = workoutSession?.associatedWorkoutBuilder()
            
            workoutSession?.delegate = self
            workoutBuilder?.delegate = self
            
            workoutBuilder?.dataSource = HKLiveWorkoutDataSource(healthStore: healthStore,
                                                                workoutConfiguration: configuration)
            
            workoutSession?.startActivity(with: Date())
            workoutBuilder?.beginCollection(withStart: Date()) { [weak self] success, error in
                DispatchQueue.main.async {
                    if success {
                        self?.isTrackingWorkout = true
                    }
                    
                    if let error = error {
                        print("Workout start error: \(error)")
                    }
                }
            }
            
        } catch {
            print("Failed to start workout: \(error)")
        }
    }
    
    func pauseWorkout() {
        workoutSession?.pause()
    }
    
    func resumeWorkout() {
        workoutSession?.resume()
    }
    
    func stopWorkout() {
        workoutSession?.end()
        workoutBuilder?.endCollection(withEnd: Date()) { [weak self] success, error in
            if let error = error {
                print("Workout end error: \(error)")
                return
            }
            
            self?.workoutBuilder?.finishWorkout { workout, error in
                DispatchQueue.main.async {
                    self?.isTrackingWorkout = false
                    self?.currentWorkout = workout
                    self?.workoutSession = nil
                    self?.workoutBuilder = nil
                    
                    if let error = error {
                        print("Workout finish error: \(error)")
                    }
                }
            }
        }
    }
    
    // MARK: - Health Milestones and Alerts
    
    private func checkStepMilestones(_ steps: Int) {
        let milestones = [1000, 2500, 5000, 7500, 10000, 12500, 15000]
        
        for milestone in milestones {
            if steps == milestone {
                WatchNotificationManager.shared.addNotification(
                    id: "steps_milestone_\(milestone)",
                    type: .system,
                    title: "步数里程碑",
                    message: "恭喜！您已完成 \(milestone) 步"
                )
                
                // Trigger haptic feedback
                WKInterfaceDevice.current().play(.success)
                break
            }
        }
    }
    
    private func checkCalorieMilestones(_ calories: Int) {
        let milestones = [100, 250, 500, 750, 1000]
        
        for milestone in milestones {
            if calories == milestone {
                WatchNotificationManager.shared.addNotification(
                    id: "calories_milestone_\(milestone)",
                    type: .system,
                    title: "卡路里里程碑",
                    message: "恭喜！您已燃烧 \(milestone) 卡路里"
                )
                
                // Trigger haptic feedback
                WKInterfaceDevice.current().play(.success)
                break
            }
        }
    }
    
    private func checkHeartRateAlerts(_ heartRate: Int) {
        switch heartRate {
        case 180...:
            WatchNotificationManager.shared.addNotification(
                id: "heart_rate_high",
                type: .system,
                title: "心率过高",
                message: "当前心率 \(heartRate) BPM，请注意休息"
            )
            WKInterfaceDevice.current().play(.failure)
            
        case ...50:
            WatchNotificationManager.shared.addNotification(
                id: "heart_rate_low",
                type: .system,
                title: "心率过低",
                message: "当前心率 \(heartRate) BPM，如有不适请咨询医生"
            )
            WKInterfaceDevice.current().play(.failure)
            
        default:
            break
        }
    }
    
    // MARK: - Task Integration
    
    func getTaskHealthBonus(steps: Int, calories: Int, heartRate: Int?) -> Int {
        var bonus = 0
        
        // Steps bonus
        switch steps {
        case 10000...:
            bonus += 50
        case 5000..<10000:
            bonus += 25
        case 2000..<5000:
            bonus += 10
        default:
            break
        }
        
        // Calories bonus
        switch calories {
        case 500...:
            bonus += 30
        case 300..<500:
            bonus += 20
        case 150..<300:
            bonus += 10
        default:
            break
        }
        
        // Heart rate bonus (if in healthy range during activity)
        if let hr = heartRate {
            switch hr {
            case 120...160:
                bonus += 20 // Active range
            case 100..<120:
                bonus += 10 // Moderate range
            default:
                break
            }
        }
        
        return bonus
    }
    
    func isHealthGoalMet() -> Bool {
        return healthMetrics.steps >= 8000 || healthMetrics.calories >= 300
    }
    
    // MARK: - Data Models
    
    struct HealthMetrics {
        var steps: Int = 0
        var calories: Int = 0
        var distance: Double = 0.0 // in meters
        var heartRate: Int? = nil
        var activeMinutes: Int = 0
        var lastUpdated: Date = Date()
    }
}

// MARK: - HKWorkoutSessionDelegate

extension WatchHealthManager: HKWorkoutSessionDelegate {
    func workoutSession(_ workoutSession: HKWorkoutSession, didChangeTo toState: HKWorkoutSessionState, from fromState: HKWorkoutSessionState, date: Date) {
        DispatchQueue.main.async {
            switch toState {
            case .running:
                print("Workout started")
            case .paused:
                print("Workout paused")
            case .ended:
                print("Workout ended")
            default:
                break
            }
        }
    }
    
    func workoutSession(_ workoutSession: HKWorkoutSession, didFailWithError error: Error) {
        print("Workout session failed: \(error)")
    }
}

// MARK: - HKLiveWorkoutBuilderDelegate

extension WatchHealthManager: HKLiveWorkoutBuilderDelegate {
    func workoutBuilder(_ workoutBuilder: HKLiveWorkoutBuilder, didCollectDataOf collectedTypes: Set<HKSampleType>) {
        // Update health metrics with live workout data
        for type in collectedTypes {
            if let quantityType = type as? HKQuantityType {
                let statistics = workoutBuilder.statistics(for: quantityType)
                
                DispatchQueue.main.async {
                    switch quantityType {
                    case HKQuantityType.quantityType(forIdentifier: .stepCount):
                        if let sum = statistics?.sumQuantity() {
                            self.healthMetrics.steps = Int(sum.doubleValue(for: .count()))
                        }
                        
                    case HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned):
                        if let sum = statistics?.sumQuantity() {
                            self.healthMetrics.calories = Int(sum.doubleValue(for: .kilocalorie()))
                        }
                        
                    case HKQuantityType.quantityType(forIdentifier: .distanceWalkingRunning):
                        if let sum = statistics?.sumQuantity() {
                            self.healthMetrics.distance = sum.doubleValue(for: .meter())
                        }
                        
                    case HKQuantityType.quantityType(forIdentifier: .heartRate):
                        if let mostRecent = statistics?.mostRecentQuantity() {
                            self.healthMetrics.heartRate = Int(mostRecent.doubleValue(for: HKUnit(from: "count/min")))
                        }
                        
                    default:
                        break
                    }
                    
                    self.healthMetrics.lastUpdated = Date()
                }
            }
        }
    }
    
    func workoutBuilderDidCollectEvent(_ workoutBuilder: HKLiveWorkoutBuilder) {
        // Handle workout events if needed
    }
}